package kv

import "time"

type Storage interface {
	Set(key, value string, ttl time.Duration)
	Get(key string) (value string, ok bool)
	Delete(key string) (ok bool)
}
