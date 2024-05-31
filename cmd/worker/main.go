package main

import (
	// "context"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rixtrayker/demo-smpp/internal/app"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/db"
	"github.com/rixtrayker/demo-smpp/internal/models"
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
	// if its empty stop the program
	// Channel to receive shutdown signals
	quit := make(chan os.Signal, 1)
	// Notify on SIGINT (CTRL+C) and SIGTERM (termination signal)
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

	myDb, err = db.GetDBInstance(ctx)
	if err != nil {
		logrus.Fatalf("failed to connect to database: %v", err)
	}
	testDbAndModels(myDb)

	wg.Add(1)
	go func() {
		defer wg.Done()
		app.StartWorker(ctx, &cfg)
	}()

	wg.Wait()

	logrus.Println("Worker stopped. Exiting.")
}

func testDbAndModels(db *gorm.DB) {
	tx := db.Create(&models.DlrSms{MessageID: "message_id", MessageState: "message_state", ErrorCode: "error_code", MobileNo: 1234567890, Data: "data"})

	if tx.Error != nil {
		// logrus.Fatalf("failed to insert data: %v", tx.Error)
		logrus.Printf("failed to insert data: %v\n", tx.Error)
	}

	logrus.Printf("inserted data: %v", tx)
	// db.AutoMigrate(&models.Number{}, &models.Number2{}, &models.NumberHlr{}, &models.NumberReport{}, &models.NumberReport2023061311_48{})
}