package storages

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/xxlv/go-pluginx/storage"
)

type RedisStorage struct {
	Client *redis.Client
}

func NewRedisStorage(addr, password string, db int) *RedisStorage {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	return &RedisStorage{
		Client: client,
	}
}

func (rs *RedisStorage) Store(kv storage.Kv) error {
	err := rs.Client.Set(context.Background(), kv.Key, kv.Value, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rs *RedisStorage) Delete(key string) error {
	err := rs.Client.Del(context.Background(), key).Err()
	if err != nil {
		return err
	}
	return nil
}

func (rs *RedisStorage) Get(key string) (string, error) {
	value, err := rs.Client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return "", errors.New("not found")
	} else if err != nil {
		return "", err
	}
	return value, nil
}
