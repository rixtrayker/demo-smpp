package main

import (
	"github.com/rixtrayker/demo-smpp/internal/config"
    "github.com/rixtrayker/demo-smpp/internal/db"
    "github.com/rixtrayker/demo-smpp/internal/smpp"
    "github.com/rixtrayker/demo-smpp/internal/app"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        logrus.Fatal("Error loading .env file")
    }

    cfg := config.LoadConfig()
    db := db.Init(cfg)

    providerSessions := map[string]*smpp.Session{
        "A": smpp.NewSession(cfg, "A"),
        "B": smpp.NewSession(cfg, "B"),
        "C": smpp.NewSession(cfg, "C"),
    }

    smppClient := smpp.NewSession(providerSessions, cfg)

    app.StartWorker(db, smppClient, cfg)
}