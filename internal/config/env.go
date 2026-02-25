package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr        string
	DefaultLimit    int
	CleanupInterval time.Duration
	BlockDuration   time.Duration
	TokenLimits     string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
}

func LoadFromEnv() (Config, error) {
	defaultLimit, err := requiredInt("RATELIMIT")
	if err != nil {
		return Config{}, err
	}

	cleanupMs, err := requiredInt("RATELIMIT_CLEANUP_INTERVAL")
	if err != nil {
		return Config{}, err
	}

	blockMs, err := requiredInt("RATELIMIT_BLOCK_TIME")
	if err != nil {
		return Config{}, err
	}

	redisAddr, err := requiredString("RATELIMIT_REDIS_URL")
	if err != nil {
		return Config{}, err
	}

	redisDB := 0
	if rawDB := os.Getenv("RATELIMIT_REDIS_DB"); rawDB != "" {
		parsedDB, parseErr := strconv.Atoi(rawDB)
		if parseErr != nil {
			return Config{}, fmt.Errorf("invalid RATELIMIT_REDIS_DB: %w", parseErr)
		}
		redisDB = parsedDB
	}

	httpAddr := os.Getenv("RATELIMIT_HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}

	return Config{
		HTTPAddr:        httpAddr,
		DefaultLimit:    defaultLimit,
		CleanupInterval: time.Millisecond * time.Duration(cleanupMs),
		BlockDuration:   time.Millisecond * time.Duration(blockMs),
		TokenLimits:     os.Getenv("RATELIMIT_TOKEN_LIST"),
		RedisAddr:       redisAddr,
		RedisPassword:   os.Getenv("RATELIMIT_REDIS_PASSWORD"),
		RedisDB:         redisDB,
	}, nil
}

func requiredInt(name string) (int, error) {
	raw, err := requiredString(name)
	if err != nil {
		return 0, err
	}

	value, parseErr := strconv.Atoi(raw)
	if parseErr != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, parseErr)
	}

	return value, nil
}

func requiredString(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("missing required environment variable: %s", name)
	}
	return value, nil
}
