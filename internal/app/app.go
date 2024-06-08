package app

import (
	"context"
	"log"
	"reflect"
	"sync"

	"github.com/rixtrayker/demo-smpp/internal/config"
	clients "github.com/rixtrayker/demo-smpp/internal/gateway"
	"github.com/sirupsen/logrus"
)

var appClients map[string]*clients.ClientBase

func initClients(ctx context.Context, cfg *config.Config) {
	appClients = make(map[string]*clients.ClientBase)
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

			client, err := clients.NewClientBase(ctx, provider, provider.Name)

			if err != nil {
				logrus.WithError(err).Error("Failed to create client for provider", provider.Name)
				return
			}

			client.Start()
			if client != nil {
				mu.Lock()
				appClients[provider.Name] = client
				mu.Unlock()
			}
		}(provider)
	}

	wg.Wait()
}

func Start(ctx context.Context, cfg *config.Config) {
	go handleShutdown(ctx)

	initClients(ctx, cfg)

	
}

func handleShutdown(ctx context.Context) {
	<-ctx.Done()
	log.Println("Context canceled, initiating shutdown...")
	for _, c := range appClients {
		c.Stop()
	}
	log.Println("All sessions stopped.")
}
