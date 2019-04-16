package common

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

type IKVStore interface {
	Set(key string, value string, expire time.Duration) error
	Get(key string) (string, error)
}

var KVStoreEmpty = errors.New("KVStoreEmpty")

func InitKVStore(config interface{}) (store IKVStore, err error) {
	switch c := config.(type) {
	case nil:
		return nil, fmt.Errorf("need Config to init KVStore")
	case *RedisKVStoreConfig:
		KVStore := &RedisKVStore{}
		err = KVStore.Init(c)

		if err != nil {
			return
		}

		return KVStore, nil
	default:
		return nil, fmt.Errorf("KVStore config is not support %v", config)
	}
}

type (
	RedisKVStore struct {
		ctx    context.Context
		client *redis.Client
	}

	RedisKVStoreConfig struct {
		Ctx    context.Context
		Client *redis.Client
	}
)

func (queue RedisKVStore) Set(key, value string, expire time.Duration) error {
	ret := queue.client.Set(key, value, expire)
	return ret.Err()
}

func (queue RedisKVStore) Get(key string) (string, error) {
	ret := queue.client.Get(key)
	res, err := ret.Result()

	if err == redis.Nil {
		return "", KVStoreEmpty
	}

	return res, err
}

func (queue *RedisKVStore) Init(config *RedisKVStoreConfig) error {
	if config.Client == nil {
		return fmt.Errorf("no redis Connection")
	}

	queue.client = config.Client
	queue.ctx = config.Ctx

	return nil
}
