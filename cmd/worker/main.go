package main

import (
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

    app.StartWorker(&cfg)
}