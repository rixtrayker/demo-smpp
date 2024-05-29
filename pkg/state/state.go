package state

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

type State struct {
	mutex     *sync.RWMutex
	ctx       context.Context
	cancelFn  context.CancelFunc
	cancels   map[string]*context.CancelFunc
	state     int
	stateTime time.Time
}

func NewState(ctx context.Context) (*State, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	ctx, cancelFn := context.WithCancel(ctx)
	return &State{
		mutex:     &sync.RWMutex{},
		ctx:       ctx,
		cancelFn:  cancelFn,
		cancels:   map[string]*context.CancelFunc{},
		state:     New,
		stateTime: time.Now(),
	}, nil
}

func (s *State) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch s.state {
	case New, Stopped, Restarting:
		s.state = Running
		s.stateTime = time.Now()
		_, cancel := context.WithCancel(s.ctx)
		s.cancels["client"] = &cancel
		log.Printf("Started at %v", s.stateTime)
		return nil
	case Running:
		return errors.New("already running")
	case Terminating:
		return errors.New("terminating, cannot start")
	default:
		return fmt.Errorf("invalid state transition from %s", stateNames[s.state])
	}
}

func (s *State) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch s.state {
	case Running:
		s.state = Stopped
		s.stateTime = time.Now()
		if cancel, ok := s.cancels["client"]; ok {
			(*cancel)()
			delete(s.cancels, "client")
		}
		log.Printf("Stopped at %v", s.stateTime)
		return nil
	case Stopped, New:
		return errors.New("already stopped")
	case Terminating:
		return errors.New("terminating, cannot stop")
	default:
		return fmt.Errorf("invalid state transition from %s", stateNames[s.state])
	}
}

func (s *State) Restart() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch s.state {
	case Running:
		s.state = Restarting
		s.stateTime = time.Now()
		if cancel, ok := s.cancels["client"]; ok {
			(*cancel)()
			delete(s.cancels, "client")
		}
		_, cancel := context.WithCancel(s.ctx)
		s.cancels["client"] = &cancel
		s.state = Running
		log.Printf("Restarted at %v", s.stateTime)
		return nil
	case Stopped, New:
		return s.Start()
	case Terminating:
		return errors.New("terminating, cannot restart")
	default:
		return fmt.Errorf("invalid state transition from %s", stateNames[s.state])
	}
}

func (s *State) Terminate() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch s.state {
	case Running, Paused, Restarting:
		s.state = Terminating
		s.stateTime = time.Now()
		if cancel, ok := s.cancels["client"]; ok {
			(*cancel)()
			delete(s.cancels, "client")
		}
		s.cancelFn() // Cancel the main context
		log.Printf("Terminating at %v", s.stateTime)
		return nil
	case Terminated:
		return errors.New("already terminated")
	default:
		return fmt.Errorf("invalid state transition from %s", stateNames[s.state])
	}
}

func (s *State) Pause() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch s.state {
	case Running:
		s.state = Paused
		s.stateTime = time.Now()
		log.Printf("Paused at %v", s.stateTime)
		return nil
	case Paused:
		return errors.New("already paused")
	case Terminating:
		return errors.New("terminating, cannot pause")
	default:
		return fmt.Errorf("invalid state transition from %s", stateNames[s.state])
	}
}

func (s *State) Resume() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch s.state {
	case Paused:
		s.state = Running
		s.stateTime = time.Now()
		log.Printf("Resumed at %v", s.stateTime)
		return nil
	case Running:
		return errors.New("already running")
	case Terminating:
		return errors.New("terminating, cannot resume")
	default:
		return fmt.Errorf("invalid state transition from %s", stateNames[s.state])
	}
}


func (s *State) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state == Running
}

func (s *State) IsPaused() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state == Paused
}

func (s *State) IsTerminating() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state == Terminating
}

func (s *State) GetState() (int, time.Time) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state, s.stateTime
}