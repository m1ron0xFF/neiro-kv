package kv

import (
	"context"
	"neiro-kv/pkg/gen/kv/v1"
)

func (s *kvServiceServer) Delete(_ context.Context, req *kv.DeleteKvRequest) (*kv.DeleteKvResponse, error) {
	ok := s.storage.Delete(req.Key)
	return &kv.DeleteKvResponse{
		Found: ok,
	}, nil
}
