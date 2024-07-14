package queue

import (
	"context"
	"sync"
	"time"

	"encoding/json"

	"github.com/phuslu/log"
	"github.com/sirupsen/logrus"

	"github.com/redis/go-redis/v9"
	"github.com/rixtrayker/demo-smpp/internal/config"

	"go.uber.org/atomic"
)

type Worker struct {
    ctx            context.Context
    sysCtx         context.Context
    cancel         context.CancelFunc
    logger         log.Logger
    redis          *redis.Client
    decoder        *Decoder
    queues         []string
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

func NewWorker(sysCtx context.Context, options ...Option) (*Worker, error) {
    decoder := NewDecoder()
    cfg := config.LoadConfig("")

    client := redis.NewClient(&redis.Options{
        PoolSize: 1000,
        Addr:     cfg.RedisURL,
        Password: "", // Ensure password is handled securely
        DB:       0,
        ReadTimeout:  5 * time.Second, // or -1 for no timeout
        WriteTimeout: 10 * time.Second,
    })
    ctx := context.Background()
    _, err := client.Ping(ctx).Result()
    if err != nil {
        logrus.WithError(err).Fatal("Failed to connect to Redis")
        return nil, err
    }

    worker := &Worker{
        ctx:            ctx,
        sysCtx:         sysCtx,
        redis:          client,
        logger:         log.Logger{
            Level: log.InfoLevel,
            Writer: &log.FileWriter{
                Filename:   "queue.log",
                // MaxSize:    50 * 1024 * 1024,
                // MaxSize:    100<<20,
                MaxBackups: 14,
                LocalTime:  false,
            },
        },
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
        w.logger.Error().Err(err).Msg("Failed to consume message from queue")
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
            case <-w.sysCtx.Done():
                w.logger.Info().Msg("Stream queue shutting down...")
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

func (w *Worker) PushPorted(message MessageData) error {
   return w.PushMessage("go-"+message.Gateway+"-ported", message)
}

func (w *Worker) PushMessage(queue string, message MessageData) error {
    w.wg.Add(1)
    defer w.wg.Done()
    var err error
    var strMsg []byte
    // encode message to json
    strMsg, err = json.Marshal(message)
    if err != nil {
        w.logger.Error().Err(err).Msg("Failed to marshal message")
        w.logger.Error().Object("message", &message).Msg("Failed to marshal ")
        return err
    }
    _, err = w.redis.RPush(w.ctx, queue, string(strMsg)).Result()
    if err != nil {
        logrus.WithError(err).Error("Failed to push message to queue")
        return err
    }
    return nil
}

func (w *Worker) Stop() {
    w.logger.Info().Msg("Worker shutdown initiated")
    w.wg.Wait()
    w.logger.Info().Msg("w.wg wait done")
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
        w.logger.Error().Err(err).Msg("Failed to close Redis client")
        return err
    }
    close(w.errors)
    return nil
}
