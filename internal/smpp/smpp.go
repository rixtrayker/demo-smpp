package smpp

import (
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/session"
)

type Client struct {
	sessions map[string]*session.Session
	config   config.Config
}

func NewClient(sessions map[string]*session.Session, cfg config.Config) *Client {
	return &Client{
		sessions: sessions,
		config:   cfg,
	}
}

func (c *Client) SendMessage(provider string, message string) error {
	if session, ok := c.sessions[provider]; ok {
		return session.Send(message)
	}
	return nil
}
