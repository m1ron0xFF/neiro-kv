package kv

import "neiro-kv/pkg/gen/kv/v1"

type kvServiceServer struct {
	kv.UnimplementedKvServiceServer
	storage Storage
}

func NewKvServiceServer(storage Storage) kv.KvServiceServer {
	return &kvServiceServer{
		storage: storage,
	}
}
