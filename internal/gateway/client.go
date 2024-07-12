package clients

import (
	"context"
	"errors"
	"sync"

	"github.com/rixtrayker/demo-smpp/internal/app/handlers"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/queue"
	"github.com/rixtrayker/demo-smpp/internal/response"
	"github.com/rixtrayker/demo-smpp/internal/session"
	"github.com/sirupsen/logrus"
)

type ClientBase struct {
	providerName string
	ctx    context.Context
	wg     sync.WaitGroup
	cfg    config.Provider
	queues []string
	cancelFunc *context.CancelFunc
	// state  *state.State
	DeliverHandler func()
	handler *handlers.Handler
	session *session.Session
	worker  *queue.Worker
}

func NewClientBase(ctx context.Context, cfg config.Provider, name string) (*ClientBase, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	var cancel context.CancelFunc
	// ctx, cancel = context.WithCancel(ctx)
	
	rw := response.NewResponseWriter()
	sess, err := session.NewSession(cfg, nil, session.WithResponseWriter(rw))
	if err != nil {
		// cancel()
		return nil, err
	}
	// state, _ := state.NewState(ctx)
	return &ClientBase{
		providerName: name,
		ctx:    ctx,
		wg:     sync.WaitGroup{},
		cancelFunc: &cancel,
		cfg: cfg,
		session: sess,
		// state: state,
	}, nil
}

func (c * ClientBase) SetHandler(handler *handlers.Handler) {
	c.handler = handler
}

func (c *ClientBase) Start() {
	// c.state.Start()
	// handlerFunc := c.handler.Handle
	err := c.session.Start(c.ctx)
	if err != nil {
		return
	}

	w, err := queue.NewWorker(c.ctx, queue.WithQueues(c.cfg.Queues...))
	c.worker = w
	if err != nil {
		return
	}

	portedFinished := make(chan struct{})
	go func(){
		defer close(portedFinished)
		c.runPorted()
	}()

	messages, errChan := w.Stream()
	go func() {
        for err := range errChan {
           logrus.WithError(err).Error("Failed to stream message in func Stream") 
        }
    }()

	// cancelled := make(chan struct{})
	// c.session.SendStreamWithCancel(c.ctx, messages, cancelled)
	c.session.SendStreamWithCancel(c.ctx, messages)
	logrus.Infof("PushMessage len: %d", len(messages))
	bgCtx := context.Background()
	// 151 msg from streams sizes
	for msg := range messages {
		w.PushMessage(bgCtx,  "go-" + c.cfg.Name + "-resend", msg)
	}
	// latest 1 msg from cancelling
	for msg := range c.session.StreamResend() {
		w.PushMessage(bgCtx,  "go-" + c.cfg.Name + "-resend", msg)
	}
	<-portedFinished
	// send close signal to unblock stopping and closing and inside Stop() it waits for running Queue calls
	w.Finished()
}

func (c *ClientBase) runPorted() {
    msg, _ := c.session.StreamPorted()
    semaphore := make(chan struct{}, 50)
    var wg sync.WaitGroup

    for m := range msg {
        semaphore <- struct{}{} // Acquire a semaphore slot
        wg.Add(1) // Increment the WaitGroup counter
        go func(m queue.MessageData) {
            defer wg.Done() // Decrement the WaitGroup counter when done
            defer func() { <-semaphore }() // Release the semaphore slot when done
            err := c.worker.PushPorted(c.ctx, m)
            if err != nil {
                logrus.WithError(err).Error("pushing failed failed sending message")
            }
        }(m)
    }

    wg.Wait() // Wait for all goroutines to finish

    close(semaphore) // Close the semaphore channel
}

func (c *ClientBase) Stop() {
	c.session.Stop()
	c.worker.Stop()
}
