package main

import (
	// "context"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/db"
	"github.com/rixtrayker/demo-smpp/internal/models"
	"github.com/rixtrayker/demo-smpp/ping"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Fatal("Error loading .env file")
	}

	cfg := config.LoadConfig()
	// if its empty stop the program
	// Channel to receive shutdown signals
	quit := make(chan os.Signal, 1)
	// Notify on SIGINT (CTRL+C) and SIGTERM (termination signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	ctx := context.Background()
	// ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-quit
		logrus.Println("Shutting down gracefully...")
		// cancel()
	}()

	myDb, err := db.GetDBInstance(ctx, cfg.DatabaseConfig)

	if err != nil {
		logrus.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close(ctx)
	testDbAndModels(myDb)

	wg.Add(1)
	go func() {
		defer wg.Done()
		ping.TestLive(ctx, &wg)
	// app.StartWorker(ctx, &cfg)
	}()

	wg.Wait()

	logrus.Println("Worker stopped. Exiting.")
}

func testDbAndModels(db *gorm.DB) {
	tx := db.Create(&models.DlrResponse{Response: "response", Company: "company"})
	if tx.Error != nil {
		// logrus.Fatalf("failed to insert data: %v", tx.Error)
		fmt.Println("failed to insert data: %v", tx.Error)
	}

	db.Create(&models.DlrSms{MessageID: "message_id", MessageState: "message_state", ErrorCode: "error_code", MobileNo: 1234567890, Data: "data"})

	logrus.Printf("inserted data: %v", tx)
	// db.AutoMigrate(&models.Number{}, &models.Number2{}, &models.NumberHlr{}, &models.NumberReport{}, &models.NumberReport2023061311_48{})
}