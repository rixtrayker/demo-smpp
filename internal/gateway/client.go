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

	w, err := queue.NewWorker(c.ctx, queue.WithQueues(c.cfg.Queues...), queue.WithPortedQueue(c.cfg.PortedQueue))
	c.worker = w
	if err != nil {
		return
	}
	// c.wg.Add(1)
	go func(){
		defer c.wg.Done()
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
	c.wg.Add(len(messages))
	c.session.SendStreamWithCancel(c.ctx, messages)
	logrus.Infof("PushMessage len: %d", len(messages))
	bgCtx := context.Background()
	for msg := range messages {
		w.PushMessage(bgCtx, "resend", msg)
		// c.wg.Done()
	}

	// send close signal to unblock stopping and closing and inside Stop() it waits for running Queue calls
	w.Finished()

	// for wait random time and check len c.session.ResendChannel then close it
}

func(c *ClientBase) runPorted(){
	msg, _ := c.session.StreamPorted()
	// make sure it gets empty
	semaphore := make(chan struct{}, 50) // Create a semaphore with a capacity of 50
	for msg := range msg {
		semaphore <- struct{}{} // Acquire a semaphore slot
		go func(m queue.MessageData) {
			defer func() { <-semaphore }() // Release the semaphore slot when done
			err := c.worker.PushPorted(c.ctx, m)
			if err != nil {
				logrus.WithError(err).Error("pushing failed failed sending message")
			}
		}(msg)
	}
	// Wait for all goroutines to finish
	for i := 0; i < cap(semaphore); i++ {
		semaphore <- struct{}{}
	}
	c.session.ClosePorted()
}

func (c *ClientBase) Stop() {
	c.session.Stop()
	// c.wg.Wait()
	c.worker.Stop()
}
