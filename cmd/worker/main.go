package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rixtrayker/demo-smpp/internal/app"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/metrics"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.LoadConfig()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-quit
		logrus.Println("Shutting down gracefully...")
		cancel()
	}()

	metrics.StartPrometheusServer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Start(ctx, cfg)
	}()

	wg.Wait()
	logrus.Println("Application shutdown complete.")
}
