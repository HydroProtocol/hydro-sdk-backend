package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

// Iqueue is an interface of common queue service
// You can use your favourite backend to handle messages.
type IQueue interface {
	Push([]byte) error

	// Pop should not block the current thread all the time.
	Pop() ([]byte, error)
}

func InitQueue(config interface{}) (queue IQueue, err error) {
	switch c := config.(type) {
	case nil:
		return nil, fmt.Errorf("Need Config to init queue")
	case *RedisQueueConfig:
		client := &RedisQueue{}
		err = client.Init(c)

		if err != nil {
			return
		}
		return client, nil
	default:
		return nil, fmt.Errorf("Config is not support %v", config)
	}
}

// Redis Queue Implement

var EXIT = errors.New("EXIT")

type (
	RedisQueue struct {
		name   string
		ctx    context.Context
		client *redis.Client
	}

	RedisQueueConfig struct {
		Name   string
		Ctx    context.Context
		Client *redis.Client
	}
)

func (queue *RedisQueue) Push(data []byte) error {
	ret := queue.client.LPush(queue.name, data)
	return ret.Err()
}

func (queue *RedisQueue) Pop() ([]byte, error) {
	for {
		select {
		case <-queue.ctx.Done():
			return nil, EXIT
		default:
			res, err := queue.client.BRPop(time.Second, queue.name).Result()

			if err == redis.Nil {
				continue
			} else if err != nil {
				return nil, err
			}

			return []byte(res[1]), err
		}
	}
}

func (queue *RedisQueue) Init(config *RedisQueueConfig) error {
	if config.Client == nil {
		return fmt.Errorf("No redis Connection")
	}

	queue.client = config.Client
	queue.ctx = config.Ctx
	queue.name = config.Name

	return nil
}
