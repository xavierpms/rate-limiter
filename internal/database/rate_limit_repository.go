package database

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/xavierpms/rate-limiter/internal/domain/ratelimit"
	"github.com/xavierpms/rate-limiter/internal/ports"
)

type RedisRateLimitRepository struct {
	client *RedisClient
}

func NewRedisRateLimitRepository(client *RedisClient) *RedisRateLimitRepository {
	return &RedisRateLimitRepository{client: client}
}

func (r *RedisRateLimitRepository) Get(ctx context.Context, key string) (ratelimit.State, error) {
	raw, err := r.client.Get(ctx, key)
	if errors.Is(err, ErrNotFound) {
		return ratelimit.State{}, ports.ErrStateNotFound
	}
	if err != nil {
		return ratelimit.State{}, err
	}

	var state ratelimit.State
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return ratelimit.State{}, err
	}

	return state, nil
}

func (r *RedisRateLimitRepository) Save(ctx context.Context, state ratelimit.State) error {
	body, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, state.Key, body)
}

func (r *RedisRateLimitRepository) ListKeys(ctx context.Context) ([]string, error) {
	return r.client.Keys(ctx, "*")
}

func (r *RedisRateLimitRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key)
}
