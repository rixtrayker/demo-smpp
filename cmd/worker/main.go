package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/phuslu/log"
	"github.com/rixtrayker/demo-smpp/internal/app"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/metrics"
	"github.com/sirupsen/logrus"
)

var logger log.Logger

func main() {
	
	logger = log.Logger{
		Level: log.InfoLevel,
		TimeFormat: "2006-01-02 15:04:05",
		Writer: &log.FileWriter{
			Filename: "run.log",
			MaxBackups: 14,
			LocalTime:  false,
		},
	}
	cfgFile := ""

	if len(os.Args) > 1 {
		cfgFile = os.Args[1]
		logger.Info().Int("pid", os.Getpid()).Strs("args", os.Args[1:]).Msg("Run")
	} else {
		logger.Info().Int("pid", os.Getpid()).Msg("Run")
	}
	// log with args
	
	err := WritePID(".PID")
	if err != nil {
		logger.Error().Err(err).Msg("Error writing PID file")
	}
	// custom cfg file path from flags main args
	//args 

	
	cfg := config.LoadConfig(cfgFile)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-quit
		logger.Info().Msg("Shutting down gracefully...")
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
	logger.Info().Msg("Application shutdown complete.")
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