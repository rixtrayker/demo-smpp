package queue

import (
	"context"
	"errors"
	"sync"

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
    errors         chan error
    rateLimitCount *atomic.Int64
    wg             sync.WaitGroup
    shutdown       chan struct{}
}

type Option func(*Worker)


func WithQueues(queues ...string) Option {
    return func(w *Worker) {
        w.queues = queues
    }
}

func NewWorker(ctx context.Context, options ...Option) (*Worker, error) {
    decoder := NewDecoder()
    cfg := config.LoadConfig()

    client := redis.NewClient(&redis.Options{
        Addr:    cfg.RedisURL,
        Password: "",
        DB:       0,
    })

    if client == nil {
        return nil, errors.New("failed to connect to redis")
    }

    // ctx, cancel := context.WithCancel(context.Background())

    worker := &Worker{
        ctx:            ctx,
        // cancel:         cancel,
        redis:          client,
        decoder:        decoder,
        errors:         make(chan error, 100),
        rateLimitCount: atomic.NewInt64(0),
        shutdown:       make(chan struct{}),
    }

    for _, opt := range options {
        opt(worker)
    }

    return worker, nil
}
func (w *Worker) Consume() (QueueMessage, error) {
	result, err := w.redis.BLPop(w.ctx, 0, w.queues...).Result()
	if err != nil {
		return QueueMessage{}, err
	}

	return w.decoder.DecodeJSON([]byte(result[1]))
}

func (w *Worker) streamQueueMessage() (<-chan QueueMessage, <-chan error) {
    messages := make(chan QueueMessage, 200)
    errors := make(chan error)

    w.wg.Add(1)
    go func() {
        defer w.wg.Done()
        defer close(messages)
        defer close(errors)

        for {
            select {
            case <-w.ctx.Done(): // Explicitly handle context done
                logrus.Info("Stream queue message shutdown initiated")
                return
            case <-w.shutdown: // Assuming there's a shutdown channel to listen to
                logrus.Info("Shutting down streamQueueMessage due to shutdown signal")
                return
            default:
                result, err := w.Consume()
                if err != nil {
                    errors <- err
                    continue
                }
                messages <- result
            }
        }
    }()

    return messages, errors
}


func (w *Worker) Stream() (<-chan MessageData, <-chan error) {
    messages, errors := w.streamQueueMessage()
    data := make(chan MessageData)
    go func() {
        defer close(data)
        for msgQ := range messages {
            for _, msg := range msgQ.Deflate() {
                data <- msg
            }
        }
    }()

    return data, errors
}

func (w *Worker) Push(ctx context.Context, queue string, message *MessageData) error {
    w.wg.Add(1)
    defer w.wg.Done()

    select {
    case <-w.ctx.Done():
        return errors.New("worker is shutting down, can't push message")
    default:
        _, err := w.redis.RPush(ctx, queue, message).Result()
        if err != nil {
            return err
        }
        return nil
    }
}

func (w *Worker) Stop() {
    close(w.shutdown)
    logrus.Info("Worker shutdown initiated")
    w.wg.Wait()
    logrus.Info("w.wg wait done")
    w.Close()
}

func (w *Worker) Close() error {
    err := w.redis.Close()
    if err != nil {
        return err
    }
    close(w.errors)
    return nil
}
