package app

import (
	"context"
	"fmt"
	"log"
	"sync"

	// "github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/config"
	// "github.com/rixtrayker/demo-smpp/internal/handlers"
	"github.com/rixtrayker/demo-smpp/internal/session"
)

func InitSessionsAndClients(ctx context.Context, cfg *config.Config) map[string]*session.Session {
    if cfg == nil {
        panic("Config is nil")
    }
    if len(cfg.ProvidersConfig) == 0 {
        log.Println("ProvidersConfig is empty")
    }

    providerSessions := make(map[string]*session.Session)
    var wg sync.WaitGroup
    mu := sync.Mutex{}

    for _, provider := range cfg.ProvidersConfig {
		if provider == (config.Provider{}) {
			log.Println("Provider " + provider.Name + " is empty")
            continue
        }
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

            // var handler func(pdu.PDU) (pdu.PDU, bool)
            // switch provider.Name {
            // case "Provider_A":
            //     handler = handlers.ProviderAHandler(providerSessions[provider.Name])
            // case "Provider_B":
            //     handler = handlers.ProviderBHandler(providerSessions[provider.Name])
            // // case "Provider_C":
            // // handler = handlers.ProviderCHandler(providerSessions[provider.Name])
            // }

            sess, err := session.NewSession(ctx, provider, nil)
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
    return providerSessions
}
func StartWorker(ctx context.Context, cfg *config.Config) {
    var sessions map[string]*session.Session
    var wg sync.WaitGroup

    select {
    case <-ctx.Done():
        log.Println("Main cancelled, stopping worker")
        for _, s := range sessions {
            s.Close()
        }
    default:
        sessions = InitSessionsAndClients(ctx, cfg)
        if len(sessions) == 0 {
            log.Fatal("No sessions to start")
            return
        }
        go func() {
            <-ctx.Done()
            log.Println("Main cancelled, stopping worker")
            for _, s := range sessions {
                s.Close()
            }
        }()

        for _, sess := range sessions {
            wg.Add(1)
            go func(s *session.Session) {
                defer wg.Done()
                test1800(ctx, &wg, s)
            }(sess)
        }
        wg.Wait()
    }
}

func test1800(ctx context.Context, wg *sync.WaitGroup, s *session.Session) {
    fmt.Println("Sending 1800 messages")
    sem := make(chan struct{}, 1000)
    for i := 0; i < 1800; i++ {
        select {
        case <-ctx.Done():
            log.Println("Stopping message sending")
            return
        default:
            wg.Add(1)
            sem <- struct{}{}
            go func(s *session.Session, i int) {
                defer wg.Done()
                defer func() { <-sem }()
                msg := fmt.Sprintf("msg %d", i)
                if err := s.Send(msg); err != nil {
                    log.Println("Failed to send message:", err)
                }
            }(s, i)
        }
    }
    for i := 0; i < cap(sem); i++ {
        sem <- struct{}{} // Wait for all goroutines to finish
    }
    close(sem)
}