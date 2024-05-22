package kv

import (
	"context"
	"neiro-kv/pkg/gen/kv/v1"
	"time"
)

func (s *kvServiceServer) Set(_ context.Context, req *kv.SetKvRequest) (*kv.SetKvResponse, error) {
	s.storage.Set(req.Key, req.Value, time.Duration(req.TtlSec)*time.Second)
	return &kv.SetKvResponse{}, nil
}
