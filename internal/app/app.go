package app

import (
	"context"
	"fmt"
	"log"

	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/smpp"
)

func InitSessionsAndClients(ctx context.Context, cfg *config.Config) (map[string]*smpp.Session, context.CancelFunc) {
	if cfg == nil {
		panic("Config is nil")
	}

	if len(cfg.ProvidersConfig) == 0 {
		log.Println("ProvidersConfig is empty")
	}

	providerSessions := make(map[string]*smpp.Session)
	ctx, cancel := context.WithCancel(ctx)

	for _, provider := range cfg.ProvidersConfig {
		if provider == (config.Provider{}) {
			log.Println("Provider is empty")
			continue
		}

		if provider.Name != "" {
			log.Println("Provider", provider.Name)
		}

		session, err := smpp.NewSession(ctx, provider)
		if err != nil {
			log.Println("Failed to create session for provider", provider.Name, err)
			continue
		}

		if session != nil {
			providerSessions[provider.Name] = session
		}
	}

	return providerSessions, cancel
}

func StartWorker(ctx context.Context, cfg *config.Config) {
	sessions, cancel := InitSessionsAndClients(ctx, cfg)
	defer cancel()

	if len(sessions) == 0 {
		log.Fatal("No sessions to start")
		return
	}

	if sessions["Provider_B"] != nil {
		test1800(ctx, sessions["Provider_B"])
	}
}

func test1800(ctx context.Context, session *smpp.Session) {
	log.Println("Sending 1800 messages")

	msg := "msg -1"
	if err := session.Send(msg); err != nil {
		log.Println("Failed to send initial message:", err)
	}

	for i := 0; i < 1800; i++ {
		select {
		case <-ctx.Done():
			log.Println("Stopping message sending")
			return
		default:
			go func(i int) {
				msg := fmt.Sprintf("msg %d", i)
				if err := session.Send(msg); err != nil {
					log.Println("Failed to send message:", err)
				}
			}(i)
		}
	}
}
