package queue

import (
	"context"
	"errors"
	"sync"

	"github.com/redis/go-redis/v9"
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

func NewWorker(queues []string) (*Worker, error) {
    decoder := NewDecoder()

    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })

    if client == nil {
        return nil, errors.New("failed to connect to redis")
    }

    ctx, cancel := context.WithCancel(context.Background())

    worker := &Worker{
        ctx:            ctx,
        cancel:         cancel,
        redis:          client,
        decoder:        decoder,
        queues:         queues,
        errors:         make(chan error, 100),
        rateLimitCount: atomic.NewInt64(0),
        shutdown:       make(chan struct{}),
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
            case <-w.ctx.Done():
                return
            case <-w.shutdown:
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
    w.wg.Wait()
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
