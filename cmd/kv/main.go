package main

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"neiro-kv/internal/server/kv"
	storage2 "neiro-kv/internal/storage"
	pbKv "neiro-kv/pkg/gen/kv/v1"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

const KVPort = 8080
const DiagHTTPPort = 8081

func main() {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	grpcKvServer := grpc.NewServer(
		grpc.StreamInterceptor(
			logging.StreamServerInterceptor(interceptorLogger(log.Default())),
		),
		grpc.UnaryInterceptor(
			logging.UnaryServerInterceptor(interceptorLogger(log.Default())),
		),
	)
	// Register reflection service on gRPC server.
	reflection.Register(grpcKvServer)

	storage := storage2.NewStorage()
	go storage.GcLoop(ctx)

	kvServer := kv.NewKvServiceServer(storage)
	pbKv.RegisterKvServiceServer(grpcKvServer, kvServer)

	// gRPC server
	go func() {
		log.Println("starting grpc server", "port", KVPort)

		listener, err := net.Listen("tcp", ":"+strconv.Itoa(KVPort))
		if err != nil {
			log.Fatalf("grpc listener run failure %s", err)
		}

		if err = grpcKvServer.Serve(listener); err != nil {
			log.Fatalf("grpc server run failure %s", err)
		}
	}()

	// Pprof API
	diagAPIRouter := chi.NewRouter()
	diagAPIRouter.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { return })
	diagAPIRouter.HandleFunc("/debug/pprof/", pprof.Index)
	diagAPIRouter.HandleFunc("/debug/pprof/profile", pprof.Profile)
	diagAPIRouter.HandleFunc("/debug/pprof/trace", pprof.Trace)
	diagAPIRouter.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	diagAPIRouter.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	diagAPIRouter.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	diagAPIRouter.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	diagAPIRouter.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
	diagAPIRouter.Handle("/debug/pprof/block", pprof.Handler("block"))
	diagAPIRouter.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
	diagServer := &http.Server{
		Addr:    ":" + strconv.Itoa(DiagHTTPPort),
		Handler: diagAPIRouter,

		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	// HTTP diag server
	go func() {
		log.Println("starting diag http server", "port", DiagHTTPPort)
		if err := diagServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("diag http server failed to start", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	defer func(sig os.Signal) {
		log.Printf("received signal %v\n", sig)
		grpcKvServer.GracefulStop()
		_ = diagServer.Shutdown(ctx)
		log.Println("goodbye")
	}(<-c)
}

// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/main/interceptors/logging/examples/log/example_test.go
func interceptorLogger(l *log.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			msg = fmt.Sprintf("DEBUG :%v", msg)
		case logging.LevelInfo:
			msg = fmt.Sprintf("INFO :%v", msg)
		case logging.LevelWarn:
			msg = fmt.Sprintf("WARN :%v", msg)
		case logging.LevelError:
			msg = fmt.Sprintf("ERROR :%v", msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
		l.Println(append([]any{"msg", msg}, fields...))
	})
}
