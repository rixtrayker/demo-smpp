package queue

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/sirupsen/logrus"
	"go.uber.org/atomic"
)

type Worker struct {
    ctx            context.Context
    cancel         context.CancelFunc
    redis          *redis.Client
    decoder        *Decoder
    queues         []string
    portedQueue    string
    errors         chan error
    rateLimitCount *atomic.Int64
    wg             sync.WaitGroup
}

type Option func(*Worker)

func WithQueues(queues ...string) Option {
    return func(w *Worker) {
        w.queues = queues
    }
}

func WithPortedQueue(queue string) Option {
    return func(w *Worker) {
        w.portedQueue = queue
    }
}

func NewWorker(ctx context.Context, options ...Option) (*Worker, error) {
    decoder := NewDecoder()
    cfg := config.LoadConfig()

    client := redis.NewClient(&redis.Options{
        Addr:     cfg.RedisURL,
        Password: "", // Ensure password is handled securely
        DB:       0,
        ReadTimeout:  5 * time.Second, // or -1 for no timeout
        WriteTimeout: 10 * time.Second,
    })

    _, err := client.Ping(ctx).Result()
    if err != nil {
        logrus.WithError(err).Fatal("Failed to connect to Redis")
        return nil, err
    }

    worker := &Worker{
        ctx:            ctx,
        redis:          client,
        decoder:        decoder,
        errors:         make(chan error, 100),
        rateLimitCount: atomic.NewInt64(0),
    }

    for _, opt := range options {
        opt(worker)
    }

    return worker, nil
}

func (w *Worker) Consume() (QueueMessage, error) {
    result, err := w.redis.BLPop(w.ctx, 1*time.Second, w.queues...).Result()
    
    if err != nil {
        if err == redis.Nil {
            return QueueMessage{}, nil
        }
        logrus.WithError(err).Error("Failed to consume message from queue")
        return QueueMessage{}, err
    }

    return w.decoder.DecodeJSON([]byte(result[1]))
}

func (w *Worker) streamQueueMessage() (<-chan QueueMessage, <-chan error) {
    messages := make(chan QueueMessage, 200)
    errChan := make(chan error,200)

    w.wg.Add(1)
    go func() {
        defer close(messages)
        defer close(errChan)
        defer w.wg.Done()

        for {
            select {
            case <-w.ctx.Done(): // Explicitly handle context done
                logrus.Info("Stream queue shutting down...")
                return
            default:
                result, err := w.Consume()
                if err != nil {
                    errChan <- err
                    continue
                }
                messages <- result
            }
        }
    }()

    return messages, errChan
}


func (w *Worker) Stream() (<-chan MessageData, <-chan error) {
    messages, errChan := w.streamQueueMessage()
    data := make(chan MessageData, 10000)
    w.wg.Add(1)
    go func() {
        defer w.wg.Done()
        defer close(data)
        for msgQ := range messages {
            for _, msg := range msgQ.Deflate() {
                data <- msg
            }
        }
    }()
    return data, errChan
}

func (w *Worker) Push(ctx context.Context, message *MessageData) error {
    w.wg.Add(1)
    defer w.wg.Done()

    select {
    case <-w.ctx.Done():
        return errors.New("worker is shutting down, can't push message")
    default:
        _, err := w.redis.RPush(ctx, w.portedQueue, message).Result()
        if err != nil {
            return err
        }
        return nil
    }
}

func (w *Worker) Stop() {
    logrus.Info("Worker shutdown initiated")
    w.wg.Wait()
    logrus.Info("w.wg wait done")
    w.Close()
}

func (w *Worker) Close() error {
    err := w.redis.Close()
    if err != nil {
        logrus.WithError(err).Error("Failed to close Redis client")
        return err
    }
    close(w.errors)
    return nil
}
