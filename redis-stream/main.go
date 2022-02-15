package rstream

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/lyineee/go-learn/utils/log"
)

type StreamConfig struct {
	Client *redis.Client
	Stream string
	logger *log.Logger
}
type RedisStream struct {
	client *redis.Client
	stream string
	logger *log.Logger
}

type GroupConfig struct {
	Group      string
	ConsumerID string
	StreamConfig
}

type ConsumerGroup struct {
	RedisStream
	group      string
	ConsumerID string
}

type XMessage struct {
	redis.XMessage
	Error error
}

func (stream *RedisStream) New(config *StreamConfig) (err error) {
	stream.client = config.Client
	stream.stream = config.Stream
	if config.logger != nil {
		stream.logger = config.logger
	} else {
		stream.logger = log.Default()
	}
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

func (stream *RedisStream) Add(ctx context.Context, value map[string]interface{}) error {
	result, err := stream.client.XAdd(ctx, &redis.XAddArgs{
		Stream:     stream.stream,
		NoMkStream: false,
		ID:         "*",
		Values:     value,
	}).Result()
	if err != nil {
		stream.logger.Error("error", log.Error(err))
	}
	if result == "0" { //TODO validate result value
		stream.logger.Info("result is 0")
	}
	return nil
}

//constructor
func (group *ConsumerGroup) New(config *GroupConfig) (err error) {
	group.client = config.Client
	group.group = config.Group
	group.stream = config.Stream
	if config.logger != nil {
		group.logger = config.logger
	} else {
		group.logger = log.Default()
	}
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
	if err != nil {
		return
	}
	err = group.CreateGroup(ctx) //check group
	if err != nil {
		return
	}
	return
}

func (group *ConsumerGroup) CreateGroup(ctx context.Context) error {
	groupInfos, err := group.client.XInfoGroups(ctx, group.stream).Result()
	if err == nil {
		for _, info := range groupInfos {
			if info.Name == group.group {
				return nil
			}
		}
	}
	group.logger.Info(fmt.Sprintf("group %s not found in stream %s", group.group, group.stream))
	result, err := group.client.XGroupCreateMkStream(ctx, group.stream, group.group, "$").Result()
	if err != nil {
		group.logger.Error("error when create group", log.Any("group_options", group.group), log.Error(err))
		return err
	}
	group.logger.Info("create group", log.String("group_option", group.group), log.String("return_code", result))
	return nil
}

func (group *ConsumerGroup) Subscribe() (c chan XMessage) {
	c = make(chan XMessage)
	go func(c chan XMessage) {
		for {
			msgs, err := group.Get(1)
			if err != nil {
				group.logger.Error("read redis group fail", log.Error(err))
				c <- XMessage{Error: err}
			}
			for _, msg := range msgs {
				//group.logger.Debug("get message", log.Any("message", msg))
				c <- XMessage{XMessage: msg}
			}
		}
	}(c)
	return
}

func (group *ConsumerGroup) Get(count int64) (message []redis.XMessage, err error) {
	ctx := context.Background()
	group.logger.Info("waiting for group message", log.String("stream", group.stream), log.String("group", group.group))
	stream, err := group.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group.group,
		Streams:  []string{group.stream, ">"},
		Consumer: group.ConsumerID,
		Count:    count,
		Block:    0,
		NoAck:    false,
	}).Result()
	if err != nil {
		return
	}
	message = stream[0].Messages
	return
}

func (group *ConsumerGroup) Pop() (message redis.XMessage, err error) {
	msgs, err := group.Get(1)
	if err != nil {
		group.logger.Error("read redis group fail", log.Error(err))
		return
	}
	message = msgs[0]
	return
}

func (group *ConsumerGroup) Ack(ctx context.Context, id string) (err error) {
	result, err := group.client.XAck(ctx, group.stream, group.group, id).Result()
	if err != nil {
		group.logger.Error("fail to ack redis queue", log.Error(err))
		return
	}
	if result == 0 {
		group.logger.Info("ack already done", log.String("message_id", id))
	}
	return
}
