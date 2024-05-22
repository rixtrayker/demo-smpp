package app

import (
	"context"
	"log"
	"gorm.io/gorm"
	"github.com/go-redis/redis/v9"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/smpp"
)

func StartWorker(db *gorm.DB, smppClient *smpp.Client, cfg config.Config) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})

	rateLimiter := config.NewRateLimiter(cfg.RateLimit)
	ctx := context.Background()

	for {
		rateLimiter.Wait(ctx)

		message, err := rdb.BLPop(ctx, 0, "go-smpp").Result()
		if err != nil {
			log.Println("Error reading from Redis:", err)
			continue
		}

		// Assume message[1] contains the actual message, message[0] contains the queue name
		err = processMessage(smppClient, message[1])
		if err != nil {
			log.Println("Error processing message:", err)
		}
	}
}

func processMessage(smppClient *smpp.Client, message string) error {
	// Determine provider based on message or other logic
	provider := "A" // Example, should be determined dynamically

	switch provider {
	case "A":
		return HandleProviderA(smppClient, message)
	case "B":
		return HandleProviderB(smppClient, message)
	default:
		return nil
	}
}
