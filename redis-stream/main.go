package rstream

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type StreamConfig struct {
	Client *redis.Client
	Stream string
}
type RedisStream struct {
	StreamConfig
	logger *zap.Logger
}

type GroupConfig struct {
	Group      string
	ConsumerID string
	StreamConfig
}

type ConsumerGroup struct {
	GroupConfig
	logger *zap.Logger
}

func (stream *RedisStream) New(config *StreamConfig) (err error) {
	stream.Client = config.Client
	stream.Stream = config.Stream
	logger, _ := zap.NewProduction()
	stream.logger = logger
	return
}

func NewStream(ctx context.Context, redisAddress, password, streamName string) (stream RedisStream, err error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: password, // no password: ""
		DB:       0,        // use default DB
	})
	config := &StreamConfig{
		Client: rdb,
		Stream: streamName,
	}
	err = stream.New(config)
	if err != nil {
		return
	}
	return
}

func (stream *RedisStream) XAdd(ctx context.Context, value map[string]interface{}) error {
	result, err := stream.Client.XAdd(ctx, &redis.XAddArgs{
		Stream:     stream.Stream,
		NoMkStream: false,
		ID:         "*",
		Values:     value,
	}).Result()
	if err != nil {
		stream.logger.Error("error", zap.Error(err))
	}
	if result == "0" { //TODO validate result value
		stream.logger.Warn("result is 0")
	}
	return nil
}

//constructor
func (group *ConsumerGroup) New(config *GroupConfig) (err error) {
	group.Client = config.Client
	group.Group = config.Group
	group.Stream = config.Stream
	logger, _ := zap.NewProduction()
	group.logger = logger
	if group.ConsumerID = config.ConsumerID; group.ConsumerID == "" {
		group.ConsumerID = uuid.NewString()

	}
	return nil
}

func NewGroup(ctx context.Context, redisAddress, password, groupName, streamName string) (group ConsumerGroup, err error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: password, // no password set
		DB:       0,        // use default DB
	})
	config := &GroupConfig{
		Group: groupName,
		StreamConfig: StreamConfig{
			Client: rdb,
			Stream: streamName,
		},
	}
	err = group.New(config)
	group.CreateGroup(ctx) //check group
	if err != nil {
		return
	}
	return
}

func (group *ConsumerGroup) CreateGroup(ctx context.Context) error {
	groupInfos, err := group.Client.XInfoGroups(ctx, group.Stream).Result()
	if err != nil {
		return err
	}
	for _, info := range groupInfos {
		if info.Name == group.Group {
			return nil
		}
	}
	group.logger.Info(fmt.Sprintf("group %s not found in stream %s", group.Group, group.Stream))
	result, err := group.Client.XGroupCreate(ctx, group.Stream, group.Group, "$ MKSTREAM").Result()
	if err != nil {
		group.logger.Error("error when create group", zap.Any("group_options", group.Group), zap.Error(err))
		return err
	}
	group.logger.Info("create group", zap.String("group_option", group.Group), zap.String("return_code", result))
	return nil
}

func (group *ConsumerGroup) Subscribe() (c chan redis.XMessage) {
	c = make(chan redis.XMessage)
	ctx := context.Background()
	group.logger.Info("waiting for group message", zap.String("stream", group.Stream), zap.String("group", group.Group))
	stream, err := group.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group.Group,
		Streams:  []string{group.Stream, ">"},
		Consumer: group.ConsumerID,
		Count:    1,
		Block:    0,
		NoAck:    false,
	}).Result()
	if err != nil {
		group.logger.Error("read redis group fail", zap.Error(err))
		return
	}
	go func(c chan redis.XMessage, stream []redis.XStream, logger *zap.Logger) {
		for _, msg := range stream[0].Messages {
			group.logger.Debug("get message", zap.Any("message", msg))
			c <- msg
		}
	}(c, stream, group.logger)
	return
}