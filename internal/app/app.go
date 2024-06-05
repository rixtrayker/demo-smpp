package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	// "github.com/linxGnu/gosmpp/pdu"

	"github.com/redis/go-redis/v9"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/queue"
	"github.com/rixtrayker/demo-smpp/internal/response"
	"github.com/sirupsen/logrus"

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

            rw := response.NewResponseWriter()
            sess, err := session.NewSession(provider, nil, session.WithResponseWriter(rw))
            if err != nil {
                logrus.WithError(err).Error("Failed to create session for provider", provider.Name)
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
    var wg sync.WaitGroup

    wg.Add(1)
    go func() {
        <-ctx.Done()
        log.Println("Main cancelled, stopping app, ctx.Done()")
        for _, s := range appSessions {
            s.Stop()
        }
        wg.Done()
    }()

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

    // Used to test concurrency

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
    fmt.Println("Sending 2000 messages")
    sem := make(chan struct{}, 1000)
    for i := 0; i < 2000; i++ {
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
                msgData := queue.MessageData{
                    Sender:      "sender",
                    Number:      "number",
                    Text:        msg,
                    Gateway:     "gateway",
                    GatewayHistory: []string{},
                }
                
                if err := s.Send(msgData); err != nil {
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

type queueMessage struct {
    Gateway string `json:"gateway"`
    PhoneNumber string    `json:"phone_number"`
    Text string `json:"text"`
}

func TestRedisConnection(ctx context.Context, client *redis.Client) error {
    if err := client.Ping(ctx).Err(); err != nil {
      return fmt.Errorf("failed to connect to redis: %w", err)
    }
    fmt.Println("Successfully connected to redis")
    return nil
}
  
func Test_redis(ctx context.Context, wg *sync.WaitGroup) error {
    // handle cancel with wg.Done()
    go func() {
        defer wg.Done()
        <-ctx.Done()
        fmt.Println("Stopping redis app")
    }()

    defer wg.Done()

    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })
    defer client.Close()

    // Close the client on exit
    if err := TestRedisConnection(ctx, client); err != nil {
        return err
    }

    queueName := "queue:go-queue-testing"

    result, err := client.BLPop(ctx, 10*time.Second, queueName).Result()
    if err != nil {
        return fmt.Errorf("failed to get message from queue: %w", err)
    }

    jsonData := result[1]

    var message queueMessage
    err = json.Unmarshal([]byte(jsonData), &message)
    if err != nil {
        return fmt.Errorf("failed to unmarshal message: %w", err)
    }

    fmt.Printf("Gateway: %s, Phone Number: %s, Text: %s\n", message.Gateway, message.PhoneNumber, message.Text)

    return nil
}

func ExampleClient() *redis.Client {
    url := "redis://user:password@localhost:6379/0?protocol=3"
    opts, err := redis.ParseURL(url)
    if err != nil {
        fmt.Println("Failed to parse URL")
        panic(err)
    }

    return redis.NewClient(opts)
}