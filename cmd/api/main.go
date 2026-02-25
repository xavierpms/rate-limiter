package main

import (
	"context"
	"log"

	"github.com/xavierpms/rate-limiter/internal/app"
	"github.com/xavierpms/rate-limiter/internal/config"
)

var loadConfig = config.LoadFromEnv
var runApplication = app.Run

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	return runApplication(context.Background(), cfg)
}
