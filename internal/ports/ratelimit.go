package ports

import (
	"context"
	"errors"

	"github.com/xavierpms/rate-limiter/internal/domain/ratelimit"
)

var ErrStateNotFound = errors.New("rate limit state not found")

type RateLimitStateRepository interface {
	Get(ctx context.Context, key string) (ratelimit.State, error)
	Save(ctx context.Context, state ratelimit.State) error
	ListKeys(ctx context.Context) ([]string, error)
	Delete(ctx context.Context, key string) error
}

type TokenLimitProvider interface {
	LimitFor(token string) int
}
