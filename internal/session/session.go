package session

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/rixtrayker/demo-smpp/internal/response"
	"github.com/sirupsen/logrus"
)

type Session struct {
    transceiver    *gosmpp.Session
    maxOutstanding int
    hasOutstanding bool
    outstandingCh  chan struct{}
    mu             sync.Mutex
    handler        func(pdu.PDU) (pdu.PDU, bool)
    stop           chan struct{}
    concatenated   map[uint8][]string
    Status         map[int32]*MessageStatus
    maxRetries     int
    responseWriter *response.Writer
    wg             sync.WaitGroup
    notCancelled   bool
}

type MessageStatus struct {
    DreamsMessageID int
    MessageID       string
    Status          string
    Number          string
}

type Option func(*Session)

func WithMaxOutstanding(maxOutstanding int) Option {
    return func(s *Session) {
        s.maxOutstanding = maxOutstanding
    }
}

func WithHasOutstanding(hasOutstanding bool) Option {
    return func(s *Session) {
        s.hasOutstanding = hasOutstanding
    }
}

func WithMaxRetries(maxRetries int) Option {
    return func(s *Session) {
        s.maxRetries = maxRetries
    }
}

func WithResponseWriter(responseWriter *response.Writer) Option {
    return func(s *Session) {
        s.responseWriter = responseWriter
    }
}

func NewSession(cfg config.Provider, handler func(pdu.PDU) (pdu.PDU, bool), options ...Option) *Session {
    session := &Session{
        concatenated:   make(map[uint8][]string),
        handler:        handler,
        outstandingCh:  make(chan struct{}, cfg.MaxOutStanding),
        stop:           make(chan struct{}),
        Status:         make(map[int32]*MessageStatus),
        wg:             sync.WaitGroup{},
        notCancelled:   true,
    }

    for _, option := range options {
        option(session)
    }

    return session
}

func (s *Session) StartSession(cfg config.Provider) error {
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
        case <-s.stop:
            return fmt.Errorf("session stopped")
        default:
            session, err := gosmpp.NewSession(
                gosmpp.TRXConnector(gosmpp.NonTLSDialer, auth),
                gosmpp.Settings{
                    EnquireLink:  5 * time.Second,
                    ReadTimeout:  10 * time.Second,
                    OnSubmitError: s.handleSubmitError,
                    OnReceivingError: s.handleReceivingError,
                    OnRebindingError: s.handleRebindingError,
                    // OnAllPDU: s.handler,
                    OnAllPDU: handlePDU(s),

                    OnClosed: s.handleClosed,
                },
                5*time.Second,
            )

            if err == nil {
                s.transceiver = session
                return nil
            }

            delay := calculateBackoff(initialDelay, maxDelay, factor, retries)
            logrus.WithError(err).Errorf("session.go:Failed to create session for provider %s", cfg.Name)

            time.Sleep(delay)
        }
    }

    return fmt.Errorf("failed to create session after %d retries", s.maxRetries)
}

func (s *Session) handleSubmitError(p pdu.PDU, err error) {
    log.Println("SubmitPDU error:", err)
}

func (s *Session) handleReceivingError(err error) {
    fmt.Println("Receiving PDU/Network error:", err)
}

func (s *Session) handleRebindingError(err error) {
    fmt.Println("Rebinding error:", err)
}

func (s *Session) handleClosed(state gosmpp.State) {
    fmt.Println(state)
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
            if s.hasOutstanding {
                <-s.outstandingCh
            }
            msgID := pd.MessageID
            ref := pd.SequenceNumber
            s.mu.Lock()
            s.Status[ref].MessageID = msgID
            s.Status[ref].Status = "pending"
            s.responseWriter.WriteResponse(&dtos.ReceiveLog{
                MessageID:    msgID,
                MobileNo:     s.Status[ref].Number,
                MessageState: "SubmitSMResp",
            })
            s.mu.Unlock()
            logrus.Info("SubmitSMResp Received")
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

func (s *Session) Stop() {
    close(s.outstandingCh)
    for len(s.outstandingCh) > 0 {
    }
    close(s.stop)
    if s.transceiver != nil {
        s.transceiver.Close()
    }
    s.wg.Wait()
    s.responseWriter.Close()
}