package clients

import (
	"context"
	"errors"
	"strconv"

	"github.com/rixtrayker/demo-smpp/internal/app/handlers"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/queue"
	"github.com/rixtrayker/demo-smpp/internal/response"
	"github.com/rixtrayker/demo-smpp/internal/session"
	"github.com/rixtrayker/demo-smpp/pkg/state"
	"github.com/sirupsen/logrus"
)

type ClientBase struct {
	ctx    context.Context
	cancel context.CancelFunc
	cfg	config.Provider
	handler *handlers.Handler
	state  *state.State
	DeliverHandler func()
}



func NewClientBase(ctx context.Context) (*ClientBase, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	ctx, cancel := context.WithCancel(ctx)
	state, _ := state.NewState(ctx)
	return &ClientBase{
		ctx:    ctx,
		cancel: cancel,
		state: state,
	}, nil
}

func (c * ClientBase) SetConfig(cfg config.Provider) {
	c.cfg = cfg
}

func (c * ClientBase) SetHandler(handler *handlers.Handler) {
	c.handler = handler
}

func (c *ClientBase) Start() {
	c.state.Start()
	// handlerFunc := c.handler.Handle
	rw := response.NewResponseWriter(c.ctx)
	sess := session.NewSession(c.cfg, nil, rw)
	err := sess.StartSession(c.cfg)
	if err != nil {
		return
	}
	defer sess.Stop()

	w, err := queue.NewWorker()
	if err != nil {
		return
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			msg, err := w.Consume(c.ctx)
			if err != nil {
				logrus.WithError(err).Error("failed to consume message")
				continue
			}
			for _, number := range msg.PhoneNumbers {
				go func(number int64){
					phoneNumber := strconv.FormatInt(number, 10)
					err = sess.Send(msg.Sender, msg.Text, phoneNumber)
				}(number)
			}

			if err != nil {
				logrus.WithError(err).Error("failed to send message")
			}
		}
	}
}

func (c *ClientBase) Stop() {
	c.state.Stop()
}

