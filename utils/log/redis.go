package log

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisWriter struct {
	client *redis.Client
	stream string
	label  string
}

func NewRedisWriter(rdb *redis.Client, stream, label string) *RedisWriter {
	return &RedisWriter{
		client: rdb,
		stream: stream,
		label:  label,
	}
}

func NewRedisWriterWithAddress(address, password, stream, label string) *RedisWriter {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
	})
	return NewRedisWriter(rdb, stream, label)
}

func (w *RedisWriter) Write(p []byte) (int, error) {
	contextTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := w.client.XAdd(contextTimeout, &redis.XAddArgs{
		Stream: w.stream,
		Values: map[string]string{w.label: string(p)},
	}).Result()
	return len(p), err
}
