package session

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/sirupsen/logrus"
)

type Session struct {
	transceiver   *gosmpp.Session
	ctx           context.Context
	maxOutstanding int
	outstandingCh chan struct{}
	mu            sync.Mutex
	handler       func(pdu.PDU) (pdu.PDU, bool)
	concatenated map[uint8][]string
	Status map[int32]*MessageStatus
	maxRetries    int
}

type MessageStatus struct {
	DreamsMessageID int
	MessageID string
	Status string
	Number string
}

func NewSession(ctx context.Context,cfg config.Provider, handler func(pdu.PDU) (pdu.PDU, bool)) (*Session, error) {
	session := &Session{
		ctx:          ctx,
		concatenated: make(map[uint8][]string),
		handler:      handler,
		maxRetries:     cfg.MaxRetries,
		maxOutstanding: cfg.MaxOutStanding,
        outstandingCh:  make(chan struct{}, cfg.MaxOutStanding),
		Status: make(map[int32]*MessageStatus),
	}
	err := session.createSession(cfg)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Session) createSession(cfg config.Provider) error {
    auth := gosmpp.Auth{
        SMSC:       cfg.SMSC,
        SystemID:   cfg.SystemID,
        Password:   cfg.Password,
        SystemType: "",
    }

    initialDelay := 100 * time.Millisecond
    maxDelay := 10 * time.Second
    factor := 2.0

    for retries := 0; retries <= s.maxRetries; retries++ {
        select {
        case <-s.ctx.Done():
            return fmt.Errorf("context cancelled during session creation")
        default:
            trans, err := gosmpp.NewSession(
                gosmpp.TRXConnector(gosmpp.NonTLSDialer, auth),
                gosmpp.Settings{
                    EnquireLink: 5 * time.Second,
					ReadTimeout: 10 * time.Second,

					OnSubmitError: func(_ pdu.PDU, err error) {
						log.Println("SubmitPDU error:", err)
					},

					OnReceivingError: func(err error) {
						fmt.Println("Receiving PDU/Network error:", err)
					},

					OnRebindingError: func(err error) {
						fmt.Println("Rebinding error:", err)
					},

					OnAllPDU: handlePDU(s),

					OnClosed: func(state gosmpp.State) {
						fmt.Println(state)
					},
				},
				5*time.Second,
            )

            if err == nil {
                s.transceiver = trans
                return nil
            }

            delay := calculateBackoff(initialDelay, maxDelay, factor, retries)
			logrus.WithError(err).Errorf("Failed to create session for provider %s", cfg.Name)
            if strings.Contains(err.Error(), "connect: connection refused") {
				logrus.WithError(err).Errorf("Connection refused. Ensure the SMSC server is running and accessible on %s", cfg.SMSC)
            }

            time.Sleep(delay)
        }
    }

	return fmt.Errorf("failed to create session after %d retries", s.maxRetries)
}

func handlePDU(s *Session) func(pdu.PDU) (pdu.PDU, bool) {
    return func(p pdu.PDU) (pdu.PDU, bool) {
        switch pd := p.(type) {
        case *pdu.Unbind:
            fmt.Println("Unbind Received")
            return pd.GetResponse(), true
        case *pdu.UnbindResp:
            // fmt.Println("UnbindResp Received")
        case *pdu.SubmitSMResp:
            <-s.outstandingCh
			msgID := pd.MessageID
			ref := pd.SequenceNumber
			s.mu.Lock()
			defer s.mu.Unlock()
			s.Status[ref].MessageID = msgID
			s.Status[ref].Status = "pending"
            // fmt.Println("SubmitSMResp Received")
        case *pdu.GenericNack:
            // fmt.Println("GenericNack Received")
        case *pdu.EnquireLinkResp:
            // fmt.Println("EnquireLinkResp Received")
        case *pdu.EnquireLink:
            // fmt.Println("EnquireLink Received")
            return pd.GetResponse(), false
        case *pdu.DataSM:
            // fmt.Println("DataSM Received")
            return pd.GetResponse(), false
        case *pdu.DeliverSM:
			return s.handleDeliverSM(pd)
        }
        return nil, false
    }
}

func (s *Session) Close() {
    s.mu.Lock()
    defer s.mu.Unlock()

    for i := 0; i < cap(s.outstandingCh); i++ {
        s.outstandingCh <- struct{}{}
    }

    close(s.outstandingCh)
    s.transceiver.Close()
}