package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/xavierpms/rate-limiter/internal/config"
	"github.com/xavierpms/rate-limiter/internal/database"
	"github.com/xavierpms/rate-limiter/internal/usecase"
	"github.com/xavierpms/rate-limiter/internal/web/handler"
	"github.com/xavierpms/rate-limiter/internal/web/middleware"
)

func NewHTTPHandler(limiter middleware.Limiter) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", handler.HelloWorldHandler)
	return middleware.RateLimitMiddleware(mux, limiter)
}

func Run(ctx context.Context, cfg config.Config) error {
	tokenLimits := database.NewTokenLimitList(cfg.TokenLimits)

	redisClient, err := database.NewRedisClient(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		return fmt.Errorf("redis client error: %w", err)
	}

	repository := database.NewRedisRateLimitRepository(&redisClient)
	limiter := usecase.NewIpRateLimiter(
		ctx,
		cfg.DefaultLimit,
		cfg.CleanupInterval,
		cfg.BlockDuration,
		&tokenLimits,
		repository,
	)

	handler := NewHTTPHandler(limiter)

	log.Printf("%s started", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, handler); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}
