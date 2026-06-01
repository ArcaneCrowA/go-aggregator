package main

import (
	"github.com/ArcaneCrowA/go-aggregator/internal/app"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Error("could not load .env", zap.Error(err))
	}

	app.Start(logger)
}
