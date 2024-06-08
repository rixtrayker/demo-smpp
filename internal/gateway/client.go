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
		queues: cfg.Queues,
		// state: state,
	}, nil
}

func (c * ClientBase) SetHandler(handler *handlers.Handler) {
	c.handler = handler
}

func (c *ClientBase) Start() {
	// c.state.Start()
	// handlerFunc := c.handler.Handle
	err := c.session.Start()
	if err != nil {
		return
	}

	w, err := queue.NewWorker(c.queues)
	c.worker = w
	if err != nil {
		return
	}
	c.wg.Add(1)
	go func(){
		defer c.wg.Done()
		c.runPorted()
	}()
    messages, _ := w.Stream()

	defer c.Stop()

	c.session.SendStream(messages)
	c.wg.Wait()
	// for wait random time and check len c.session.ResendChannel then close it
}

func(c *ClientBase) runPorted(){
	msg, _ := c.session.StreamPorted()
	for msg := range msg {
		err := c.worker.Push(c.ctx, "ported", &msg)
		if err != nil {
			logrus.WithError(err).Error("pushing failed failed sending message")
		}
	}
	c.session.ClosePorted()
}

func (c *ClientBase) Stop() {
	c.session.Stop()
}
