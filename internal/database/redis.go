package database

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient(ctx context.Context, addr, password string, db int) (RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	var err error
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return RedisClient{}, err
	}

	return RedisClient{
		Client: client,
	}, nil
}

func (r *RedisClient) Get(ctx context.Context, ip string) (string, error) {
	result, err := r.Client.Get(ctx, ip).Result()
	if err == redis.Nil {
		return "", ErrNotFound
	}
	return result, err

}

func (r *RedisClient) Set(ctx context.Context, ip string, json []byte) error {
	return r.Client.Set(ctx, ip, json, 0).Err()

}

func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.Client.Keys(ctx, pattern).Result()

}

func (r *RedisClient) Del(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}
