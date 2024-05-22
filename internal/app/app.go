package app

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/rixtrayker/demo-smpp/internal/config"
	_ "github.com/rixtrayker/demo-smpp/internal/handlers"
	"github.com/rixtrayker/demo-smpp/internal/session"
)

func InitSessionsAndClients(ctx context.Context, cfg *config.Config) (map[string]*session.Session, context.CancelFunc) {
	if cfg == nil {
		panic("Config is nil")
	}

	if len(cfg.ProvidersConfig) == 0 {
		log.Println("ProvidersConfig is empty")
	}

	providerSessions := make(map[string]*session.Session)
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup

	mu := sync.Mutex{}
	for _, provider := range cfg.ProvidersConfig {
		wg.Add(1)
		go func(provider config.Provider) {
			defer wg.Done()
			if provider == (config.Provider{}) {
				log.Println("Provider is empty")
				return
			}

			if provider.Name != "" {
				log.Println("Provider", provider.Name)
			}

			sess, err := session.CreateSessionWithRetry(ctx, provider)
			if err != nil {
				log.Println("Failed to create session for provider", provider.Name, err)
				return
			}

			if sess != nil {
				mu.Lock()
				providerSessions[provider.Name] = sess
				mu.Unlock()
			}
		}(provider)
	}

	wg.Wait()

	go func() {
		<-ctx.Done()
		for _, sess := range providerSessions {
			sess.Close()
		}
	}()

	return providerSessions, cancel
}

func StartWorker(ctx context.Context, cfg *config.Config) {
	sessions, cancel := InitSessionsAndClients(ctx, cfg)
	defer cancel()

	if len(sessions) == 0 {
		log.Fatal("No sessions to start")
		return
	}

	if session, ok := sessions["Provider_B"]; ok {
		go func() {
			test1800(ctx, session)
			cancel()
		}()
	}

	<-ctx.Done()
	log.Println("Main cancelled, stopping worker")
}

func test1800(ctx context.Context, session *session.Session) {
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
