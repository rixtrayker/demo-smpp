package app

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/queue"
	"github.com/rixtrayker/demo-smpp/internal/response"
	"github.com/rixtrayker/demo-smpp/internal/session"
	"github.com/sirupsen/logrus"
)

var appSessions map[string]*session.Session
var shutdownWG sync.WaitGroup

func InitSessionsAndClients(ctx context.Context, cfg *config.Config) {
	appSessions = make(map[string]*session.Session)
	if cfg == nil {
		panic("Config is nil")
	}

	if len(cfg.ProvidersConfig) == 0 {
		log.Println("ProvidersConfig is empty")
		return
	}

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}

	for _, provider := range cfg.ProvidersConfig {
		if reflect.DeepEqual(provider, config.Provider{}) {
			log.Println("Provider " + provider.Name + " is empty")
			continue
		}
		wg.Add(1)

		go func(provider config.Provider) {
			defer wg.Done()

			if provider.Name != "" {
				log.Println("Provider", provider.Name)
			}

			rw := response.NewResponseWriter()
			sess, err := session.NewSession(provider, nil, session.WithResponseWriter(rw))
			if err != nil {
				logrus.WithError(err).Error("Failed to create session for provider", provider.Name)
				return
			}

			err = sess.Start()
			if err != nil {
				log.Println("app.go Failed to create session for provider", provider.Name, err)
				return
			}
			if sess != nil {
				mu.Lock()
				appSessions[provider.Name] = sess
				mu.Unlock()
			}
		}(provider)
	}

	wg.Wait()
}

func Start(ctx context.Context, cfg *config.Config) {
	go handleShutdown(ctx)

	InitSessionsAndClients(ctx, cfg)

	shutdownWG.Add(1)
	go func() {
		defer shutdownWG.Done()
		test1800(ctx)
	}()

	shutdownWG.Wait()
}

func handleShutdown(ctx context.Context) {
	<-ctx.Done()
	log.Println("Context canceled, initiating shutdown...")
	for _, s := range appSessions {
		s.Stop()
	}
	log.Println("All sessions stopped.")
}

func test1800(ctx context.Context) {
	fmt.Println("Sending 2000 messages")
	sem := make(chan struct{}, 1000)
	for i := 0; i < 2000; i++ {
		select {
		case <-ctx.Done():
			log.Println("Stopping message sending")
			return
		default:
			shutdownWG.Add(1)
			sem <- struct{}{}
			go func(i int) {
				defer shutdownWG.Done()
				defer func() { <-sem }()
				msg := fmt.Sprintf("msg %d", i)
				msgData := queue.MessageData{
					Sender:         "sender",
					Number:         "number",
					Text:           msg,
					Gateway:        "gateway",
					GatewayHistory: []string{},
				}

				for _, sess := range appSessions {
					if err := sess.Send(msgData); err != nil {
						log.Println("Failed to send message:", err)
					}
				}
			}(i)
		}
	}

	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}

	close(sem)
}
