package main

import (
	"github.com/rixtrayker/demo-smpp/internal/config"
    "github.com/rixtrayker/demo-smpp/internal/db"
    "github.com/rixtrayker/demo-smpp/internal/smpp"
    "github.com/rixtrayker/demo-smpp/internal/app"
    "github.com/rixtrayker/demo-smpp/internal/session"
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

    providerSessions := map[string]*session.Session{
        "A": session.New(cfg, "A"),
        "B": session.New(cfg, "B"),
        "C": session.New(cfg, "C"),
    }

    smppClient := smpp.NewClient(providerSessions, cfg)

    app.StartWorker(db, smppClient, cfg)
}
