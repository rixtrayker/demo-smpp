package queue

import (
	"context"
	"sync"
	"time"

	"encoding/json"

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
    closeCh        chan struct{}
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
        PoolSize: 1000,
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
        closeCh:       make(chan struct{}),
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
    messages := make(chan QueueMessage, 50)
    errChan := make(chan error, 50)

    w.wg.Add(1)
    go func() {
        defer close(messages)
        defer close(errChan)
        defer w.wg.Done()

        for {
            select {
            case <-w.ctx.Done():
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
    data := make(chan MessageData, 100)
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

func (w *Worker) PushPorted(ctx context.Context, message MessageData) error {
   return w.PushMessage(ctx, w.portedQueue, message)
}

func (w *Worker) PushMessage(ctx context.Context, queue string, message MessageData) error {
    w.wg.Add(1)
    defer w.wg.Done()
    var err error
    var strMsg []byte
    // encode message to json
    strMsg, err = json.Marshal(message)
    if err != nil {
        logrus.WithError(err).Error("Failed to marshal message")
        logrus.Errorf("Failed to marshal message: %v", message)
        return err
    }
    _, err = w.redis.RPush(ctx, queue, string(strMsg)).Result()
    if err != nil {
        logrus.WithError(err).Error("Failed to push message to queue")
        return err
    }
    return nil
}

func (w *Worker) Stop() {
    logrus.Info("Worker shutdown initiated")
    w.wg.Wait()
    logrus.Info("w.wg wait done")
    w.Close()
}

func (w *Worker) Finished(){
    close(w.closeCh)
}

func (w *Worker) Close() error {
    <-w.closeCh
    w.wg.Wait()
    err := w.redis.Close()
    if err != nil {
        logrus.WithError(err).Error("Failed to close Redis client")
        return err
    }
    close(w.errors)
    return nil
}
