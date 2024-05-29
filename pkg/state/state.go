package state

import (
    "context"
    "errors"
    "sync"
)

type State struct {
    mutex   *sync.Mutex
    ctx     context.Context
    cancels map[string]*context.CancelFunc
    state   int
}

const (
    Stopped = iota
    Paused
    Running
    Restarting
)

func NewState(ctx context.Context) (*State, error) {
    if ctx == nil {
        return nil, errors.New("context cannot be nil")
    }
    return &State{
        mutex:   &sync.Mutex{},
        ctx:     ctx,
        cancels: map[string]*context.CancelFunc{},
        state:   Stopped,
    }, nil
}

func (s *State) Start() {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    if s.state == Stopped {
        s.state = Running
        // Create a new context with cancellation for this specific client
        ctx, cancel := context.WithCancel(s.ctx)
        s.cancels["client"] = &cancel
    }
}

func (s *State) Stop() {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    if s.state == Running {
        s.state = Stopped
        if cancel, ok := s.cancels["client"]; ok {
            (*cancel)() // Call the cancellation function for the client context
            delete(s.cancels, "client")
        }
    }
}

func (s *State) Restart() {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    if s.state == Running {
        s.state = Restarting
        if cancel, ok := s.cancels["client"]; ok {
            (*cancel)() // Cancel the existing context before restarting
            delete(s.cancels, "client")
        }
        // Create a new context with cancellation for the restarted client
        ctx, cancel := context.WithCancel(s.ctx)
        s.cancels["client"] = &cancel
        s.state = Running
    }
}

func (s *State) IsRunning() bool {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    return s.state == Running
}

func (s *State) IsRestarting() bool {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    return s.state == Restarting
}

// WrapWork is a helper function to be embedded in client functions
// It checks the  state and cancels the client context if necessary.
func (s *State) WrapWork(fn func(context.Context)) func() error {
    return func() error {
        s.mutex.Lock()
        defer s.mutex.Unlock()

        if s.state != Running {
            return errors.New(" is not running")
        }

        ctx, cancel := context.WithCancel(s.ctx)
        defer cancel() // Ensure cancellation when the function exits

        go func() {
            select {
            case <-ctx.Done():
            case <-s.ctx.Done():
                cancel() // Cancel the client context if the  context is done
            }
        }()

        fn(ctx) // Execute the actual work function with the client context

        return nil
    }
}
