package kv

import (
	"context"
	"neiro-kv/pkg/gen/kv/v1"
)

func (s *kvServiceServer) Get(_ context.Context, req *kv.GetKvRequest) (*kv.GetKvResponse, error) {
	value, found := s.storage.Get(req.Key)
	return &kv.GetKvResponse{
		Value: value,
		Found: found,
	}, nil
}
