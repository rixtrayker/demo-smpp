package session

import "github.com/rixtrayker/demo-smpp/internal/queue"

type Stream struct {
	bufferSize 	    int
	streamClose chan struct{}
	stream      chan queue.MessageData
	err         chan error
}

func NewStream(buffer int) *Stream {
	bufferSize := 50
	if buffer > 0 {
		bufferSize = buffer
	}

	return &Stream{
		bufferSize:  bufferSize,
		streamClose: make(chan struct{}),
		stream:      make(chan queue.MessageData, bufferSize),
		err:         make(chan error),
	}
}

func (s *Stream) Wait() {
	<-s.streamClose
}

func (s *Stream) Close() {
	close(s.stream)
	close(s.streamClose)
}
