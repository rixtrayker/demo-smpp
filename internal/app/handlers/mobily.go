package handlers

import (
	"context"
	"errors"
)

type MobilyHandler struct {
	Handler
}

func NewMobilyHandler(ctx context.Context, handlerFunc func()) (*MobilyHandler, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	ctx, cancel := context.WithCancel(ctx)
	return &MobilyHandler{
		Handler: Handler{
			ctx: ctx,
			cancel: cancel,
			handlerFunc: mobilyHandlerFunc,
		},
	}, nil
}

func mobilyHandlerFunc() {
	// implement the handlerFunc
}
