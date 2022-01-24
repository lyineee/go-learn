package main

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/lyineee/go-learn/utils"
)

func initLog() {
	log := utils.GetLogger()
	logger = log.Sugar()
	defer logger.Sync()
}
func TestCreateGroup(t *testing.T) {
	initLog()
	// init redis db
	var ctx = context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	redisQueueOptions := RedisQueueOptions{
		Group:  "backend.history.refresh.workers",
		Stream: "backend.history.refresh",
	}
	err := createGroup(ctx, rdb, redisQueueOptions)
	if err != nil {
		t.Log("create group error", err)
	}
}
