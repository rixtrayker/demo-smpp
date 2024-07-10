package main

import (
	"context"
	"fmt"
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
	go WritePID(".PID")

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

func WritePID(filename string) error {
	pid := os.Getpid()
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating PID file: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%d", pid))
	if err != nil {
		return fmt.Errorf("error writing PID to file: %w", err)
	}
	return nil
}