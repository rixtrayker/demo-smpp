package ping

import (
    "context"
    "fmt"
    "os"
    "sync"
    "time"

    "github.com/rixtrayker/demo-smpp/internal/config"
    "github.com/rixtrayker/demo-smpp/internal/session"
)

var providersList []config.Provider

func TestLive(ctx context.Context, wg *sync.WaitGroup) {
    defer wg.Done()
    fmt.Println("Testing Real Numbers")
    cfg := config.LoadConfig()
    providersList = cfg.ProvidersConfig

    msg := "testing GoLang worker response"

    numbers := map[string]string{
        "Zain":   os.Getenv("ZAIN_NUMBER"),
        "Mobily": os.Getenv("MOBILY_NUMBER"),
        "STC":    os.Getenv("STC_NUMBER"),
    }

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

    fmt.Println("Testing Bassel Number")
    wg.Add(1)
    go func() {
        defer wg.Done()
        testProvider(ctx, "STC", "966551589449", msg)
    }()

    // Wait for cancellation or completion
    select {
    case <-ctx.Done():
        fmt.Println("TestLive cancelled")
    default:
        wg.Wait()
    }

    fmt.Println("TestLive completed")
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
    cfg := getProviderCfg(providerName)
    if cfg == (config.Provider{}) {
        fmt.Printf("Empty %s config\n", providerName)
        return
    }

    session, err := session.NewSession(ctx, cfg, nil)
    if err != nil {
        fmt.Println(err)
        return
    }

    err = session.Send("dreams", number, msg)
    if err != nil {
        fmt.Println("Error sending SMS:", err)
    }

    time.Sleep(30 * time.Second)
    dumpStatus(session)
}

func dumpStatus(s *session.Session) {
    for k, v := range s.Status {
        fmt.Printf("Key: %v, Value: %v\n", k, v)
    }
}
