package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rixtrayker/demo-smpp/internal/app"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env file")
	}

	cfg := config.LoadConfig()

	// Channel to receive shutdown signals
	quit := make(chan os.Signal, 1)
	// Notify on SIGINT (CTRL+C) and SIGTERM (termination signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-quit
		logrus.Println("Shutting down gracefully...")
		cancel()
	}()

	app.StartWorker(ctx, &cfg)
	logrus.Println("Worker stopped. Exiting.")
}
