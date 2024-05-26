package app

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	// "github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/config"
	// "github.com/rixtrayker/demo-smpp/internal/handlers"
	"github.com/rixtrayker/demo-smpp/internal/session"
)
var appSessions map[string]*session.Session

func InitSessionsAndClients(ctx context.Context, cfg *config.Config){
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

            sess, err := session.NewSession(ctx, provider, nil)
            if err != nil {
                log.Println("Failed to create session for provider", provider.Name, err)
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

func StartWorker(ctx context.Context, cfg *config.Config) {
    var wg sync.WaitGroup

    wg.Add(1)
    go func() {
        <-ctx.Done()
        log.Println("Main cancelled, stopping worker, ctx.Done()")
        for _, s := range appSessions {
            s.Close()
        }
        wg.Done()
    }()

    // blocking test
    // InitSessionsAndClients(ctx, cfg)

    wg.Add(1)
    go func(ctx context.Context, cfg *config.Config) {
        InitSessionsAndClients(ctx, cfg)
        wg.Done()
    }(ctx, cfg)

    time.Sleep(500 * time.Millisecond)
    
    log.Println("Starting sending messages")
    log.Println("No. of sessions: ", len(appSessions))
    for _, sess := range appSessions {
    // for i := 0; i < 5; i++ {
            wg.Add(1)
        go func(s *session.Session) {
        // go func() {
            defer wg.Done()
            // for i := 0; i < 10000000; i++ {
            //     fmt.Println("Sending message: ", i)
            // }
            test1800(ctx, &wg, s)
        // }()
        }(sess)
    }

    // Fot testing concurrenct purposes
    // for i := 0; i < 5; i++ {
    //     wg.Add(1)
    //     go func() {
    //         for i := 0; i < 10000000; i++ {
    //             select{
    //             case <-ctx.Done():
    //                 log.Println("Stopping message sending")
    //                 return
    //             default:
    //                 fmt.Println(": ", i,time.Now().UnixNano())
    //             }
    //         }
    //         wg.Done()
    //     }()
    // }

    wg.Wait()
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
        sem <- struct{}{} // bug  ??
    }

    close(sem)
}