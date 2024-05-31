package ping

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/response"
	"github.com/rixtrayker/demo-smpp/internal/session"
)

var providersList []config.Provider
var sessionsList map[string]*session.Session
var mu sync.Mutex

func TestLive(ctx context.Context) {
    cfg := config.LoadConfig()
    providersList = cfg.ProvidersConfig

    msg := "testing GoLang worker response"

    numbers := map[string]string{
        "Zain":   os.Getenv("ZAIN_NUMBER"),
        "Mobily": os.Getenv("MOBILY_NUMBER"),
        "STC":    os.Getenv("STC_NUMBER"),
    }

    var wg sync.WaitGroup
    for provider, number := range numbers {
        if number != "" {
            fmt.Printf("Testing %s Number\n", provider)
            wg.Add(1)
            go func(provider string, number string) {
                defer wg.Done()
                testProvider(ctx, provider, number, msg)
            }(provider, number)
        }
    }
    wg.Wait()

    // fmt.Println("Testing Bassel Number")
    // wg.Add(1)
    // go func() {
    //     defer wg.Done()
    //     testProvider(ctx, "STC", "966551589449", "a message from Bassel")
    // }()

    // Wait for cancellation or completion
    for _, session := range sessionsList {
        dumpStatus(session)
    }

    fmt.Println("Testing Real Numbers Done")
}

func getProviderCfg(name string) config.Provider {
    for _, provider := range providersList {
        if provider.Name == name {
            return provider
        }
    }
    return config.Provider{}
}

func testProvider(ctx context.Context, providerName, number, msg string) {
    fmt.Printf("Testing session in ping file for %s\n", providerName)

    cfg := getProviderCfg(providerName)
    if cfg == (config.Provider{}) {
        fmt.Printf("Empty %s config\n", providerName)
        return
    }


    rw := response.NewResponseWriter(ctx)
    sess := session.NewSession(cfg, nil, session.WithResponseWriter(rw))
    var err error
    select{
    case <-ctx.Done():
        sess.Stop()
        return 
    default:
        err = sess.StartSession(cfg)
        fmt.Println("try to create session")
    }
    // tried to create a session but it failed

    if err != nil || sess == nil{
        log.Println("testing_real.go Failed to create session for ", providerName, err)    
        fmt.Println(err)
        return
    }
    mu.Lock()   
    sessionsList[providerName] = sess
    mu.Unlock()

    err = sess.Send("dreams", number, msg)
    if err != nil {
        fmt.Println("Error sending SMS:", err)
    }

    time.Sleep(30 * time.Second)
}

func dumpStatus(s *session.Session) {
    for k, v := range s.Status {
        fmt.Printf("Key: %v, Value: %v\n", k, v)
    }
}
