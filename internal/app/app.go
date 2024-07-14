package app

import (
	"context"
	"reflect"
	"sync"

	"github.com/phuslu/log"
	"github.com/sirupsen/logrus"

	"github.com/rixtrayker/demo-smpp/internal/config"
	clients "github.com/rixtrayker/demo-smpp/internal/gateway"
)

var appClients map[string]*clients.ClientBase
var mu sync.Mutex = sync.Mutex{}
var wg = sync.WaitGroup{}
var logger *log.Logger

func initClients(ctx context.Context, cfg *config.Config) {
    appClients = make(map[string]*clients.ClientBase)
    if cfg == nil {
        panic("Config is nil")
    }
    if logger == nil {
        logger = &log.Logger{
            Level: log.InfoLevel,
            Caller: 1,
            TimeFormat: "15:04:05",
            Writer: &log.FileWriter{
                Filename: "app.log",
                MaxBackups: 14,
                LocalTime: false,
            },
        }
    }

    if len(cfg.ProvidersConfig) == 0 {
        logger.Error().Msg("ProvidersConfig is empty")
        logrus.Error("ProvidersConfig is empty")
        return
    }

    for _, provider := range cfg.ProvidersConfig {
        if reflect.DeepEqual(provider, config.Provider{}) {
            logger.Error().Msg("Provider " + provider.Name + " is empty")
            logrus.Error("Provider " + provider.Name + " is empty")
            continue
        }
        wg.Add(1)

        go func(provider config.Provider) {
            defer wg.Done()

            if provider.Name != "" {
                logger.Info().Msg("Provider " + provider.Name)
            }

            client, err := clients.NewClientBase(ctx, provider, provider.Name)

            if err != nil {
                logger.Error().Err(err).Msg("Failed to create client for provider " + provider.Name)
                logrus.Error("Failed to create client for provider " + provider.Name)
                return
            }

            if client != nil {
                mu.Lock()
                appClients[provider.Name] = client
                mu.Unlock()
                client.Start()
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
    logger.Info().Msg("Context canceled, initiating shutdown...")
    logrus.Info("Context canceled, initiating shutdown...")
    for _, c := range appClients {
        c.Stop()
    }
    logger.Info().Msg("All sessions stopped.")
    logrus.Info("All sessions stopped.")
}
