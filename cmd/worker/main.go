package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rixtrayker/demo-smpp/internal/app"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/db"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var myDb *gorm.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env file")
	}

	cfg := config.LoadConfig()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	go func() {
		<-quit
		logrus.Println("Shutting down gracefully...")
		cancel()
		wg.Wait()
		logrus.Println("Worker stopped. Exiting.")
	}()

	myDb, err = db.GetDBInstance()
	if err != nil {
		logrus.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if db, err := myDb.DB(); err == nil {
			db.Close()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.Start(ctx, cfg)
	}()

	wg.Wait()
}
