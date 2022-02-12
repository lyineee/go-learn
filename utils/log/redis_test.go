package log_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lyineee/go-learn/utils/log"
)

func TestRedis(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	timeout, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	_, err := rdb.Ping(timeout).Result()
	if err != nil {
		t.Error(err)
		return
	}
	writer := log.NewRedisWriter(rdb, "stream.test", "app.test")
	logger := log.NewLogger(log.NewJsonCore(writer), log.InfoLevel)

	logger.Info("test", log.Any("log", logger))
}
