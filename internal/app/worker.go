package app

import (
	"fmt"
	"log"

	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/smpp"
)

func InitSessionsAndClients(cfg *config.Config) map[string]*smpp.Session {
    if cfg == nil {
        panic("Config is nil")
    }

    if len(cfg.ProvidersConfig) == 0 {
        log.Println("ProvidersConfig is empty")
    }

    providerSessions := map[string]*smpp.Session{}

    for _, provider := range cfg.ProvidersConfig {
        if provider == (config.Provider{}) {
            log.Println("Provider is empty")
            continue
        }

        session, err := smpp.NewSession(provider)
        
        if err != nil {
            log.Println("Failed to create session for provider", provider.Name)
            continue
        }
       
        if session != nil {
            providerSessions[provider.Name] = session
        }
    }

    return providerSessions
}

func StartWorker(cfg *config.Config){
    sessions := InitSessionsAndClients(cfg)
    // for _, session := range sessions {

    // }
    if(len(sessions) == 0){
        log.Fatal("No sessions to start")
        return
    }

    test1800(sessions["Provider_B"])

}

func test1800(session *smpp.Session) {
    // send 1800
    go func() {
        for i := 0; i < 1800; i++ {
            msg := fmt.Sprintf("msg %d", i)
            session.Send(msg)
        }
    }()
}