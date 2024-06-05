package clients

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/rixtrayker/demo-smpp/internal/app/handlers"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/dtos"
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

func NewClientBase(ctx context.Context, name string) (*ClientBase, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	ctx, cancel := context.WithCancel(ctx)
	
	cfg := config.GetProviderCfg(name)
	rw := response.NewResponseWriter()
	sess := session.NewSession(cfg, nil, session.WithResponseWriter(rw))
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
	err := c.session.Start()
	if err != nil {
		return
	}

	w, err := queue.NewWorker()
	c.worker = w
	if err != nil {
		return
	}
	c.wg.Add(1)
	go func(){
		defer c.wg.Done()
		c.runPorted()
	}()
    messages, errors := w.Stream([]string{})
	// defer w.Stop()
	for {
		select {
		// case <-c.ctx.Done():
		// 	c.wg.Wait()
		// 	c.session.Close()
		// 	w.Close()
		// 	return
		case msg := <-messages:
			err := c.session.Send(msg)
			if err != nil {
				logrus.WithError(err).Error("failed sending message")
			}
		case err := <-errors:
			logrus.WithError(err).Error("failed reading from queue")
		}
	}
	
	time.Sleep(1 * time.Second)
	for len(c.session.OutstandingCh) > 0{
	}

	c.session.Stop()
	// c.wg.Wait()
	// for wait random time and check len c.session.ResendChannel then close it
}

func(c *ClientBase) runPorted(){
	msg := c.session.ResendChannel
	for msg := range msg {
		err := c.worker.Push(c.ctx, "ported", &msg)
		if err != nil {
			logrus.WithError(err).Error("failed sending message")
			c.session.Write(dtos.ReceiveLog{
				MessageID: strconv.Itoa(msg.Id),
				Gateway: msg.Gateway,
				MobileNo: msg.Number,
				MessageState: "porting failed",
				ErrorCode: err.Error(),
				Data: msg.Text,
			})
		}
	}
}

// func (c *ClientBase) Stop() {
// 	c.state.Stop()
// }
