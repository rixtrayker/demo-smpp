package session

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/queue"
	"github.com/rixtrayker/demo-smpp/internal/response"
	"github.com/sirupsen/logrus"
)

const (
    Transceiver = "transceiver"
    Receiver    = "receiver"
    Transmitter = "transmitter"
)

type Session struct {
    gateway            string
    transceiver       *gosmpp.Session
    receiver          *gosmpp.Session
    transmitter       *gosmpp.Session
    sessionType       string
    maxOutstanding    int
    hasOutstanding    bool
    OutstandingCh     chan struct{}
    ResendChannel     chan queue.MessageData
    mu                sync.Mutex
    handler           func(pdu.PDU) (pdu.PDU, bool)
    stop              chan struct{}
    concatenated      map[uint8][]string
    MessagesStatus    map[int32]*MessageStatus
    maxRetries        int
    responseWriter    *response.Writer
    wg                sync.WaitGroup
    auth              gosmpp.Auth
    EnquireLink       time.Duration
    ReadTimeout       time.Duration
    rebindingInterval time.Duration    
}

type MessageStatus struct {
    MessageID       string // submitSMResp.MessageID 
    SystemMessageID int // Dreams MessageID
    Sender          string
    Text            string
    Status          string
    Number          string
    GatewayHistory  []string
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
        gateway:        cfg.Name,
        concatenated:   make(map[uint8][]string),
        handler:        handler,
        maxOutstanding: cfg.MaxOutStanding,
        hasOutstanding: cfg.HasOutStanding,
        maxRetries:     cfg.MaxRetries,
        OutstandingCh:  make(chan struct{}, cfg.MaxOutStanding),
        stop:           make(chan struct{}),
        MessagesStatus:      make(map[int32]*MessageStatus),
        wg:             sync.WaitGroup{},
        sessionType:    cfg.SessionType,
        EnquireLink:    5 * time.Second,
        ReadTimeout:    10 * time.Second,
        rebindingInterval: 600 * time.Second,
        ResendChannel: make(chan queue.MessageData),
    }

    for _, option := range options {
        option(session)
    }

    return session
}

func (s *Session) StartSession(cfg config.Provider) error {
    s.auth = gosmpp.Auth{
        SMSC:       cfg.SMSC,
        SystemID:   cfg.SystemID,
        Password:   cfg.Password,
        SystemType: cfg.SystemType,
    }

    initialDelay := 100 * time.Millisecond
    maxDelay := 10 * time.Second
    factor := 2.0

    for retries := 0; retries <= s.maxRetries; retries++ {
        select {
        case <-s.stop:
            return fmt.Errorf("session stopped")
        default:
            err := s.sessionConnect()

            if err != nil {
                delay := calculateBackoff(initialDelay, maxDelay, factor, retries)
                logrus.WithError(err).Errorf("session.go:Failed to create session for provider %s", cfg.Name)
                time.Sleep(delay)
            } else {
                return nil
            }
        }
    }

    return fmt.Errorf("failed to create session after %d retries", s.maxRetries)
}


func (s *Session) sessionConnect() error {
    var err, err2 error
    switch s.sessionType {
    case Transceiver:   
        s.transceiver, err = gosmpp.NewSession(
            gosmpp.TRXConnector(gosmpp.NonTLSDialer, s.auth),
            s.getSetting(), 
            s.rebindingInterval,
        )
        
        if err != nil {
            return fmt.Errorf("failed to create session: %v", err)
        }
    default:
        s.receiver, err = gosmpp.NewSession(
            gosmpp.RXConnector(gosmpp.NonTLSDialer, s.auth),
            s.getSetting(),
            s.rebindingInterval,
        )

        s.transmitter, err2 = gosmpp.NewSession(
            gosmpp.TXConnector(gosmpp.NonTLSDialer, s.auth),
            s.getSetting(),
            s.rebindingInterval,
        )

        if err != nil || err2 != nil {
            return fmt.Errorf("failed to create sessions: %v, %v", err, err2)
        }
    }

    return nil
}

func (s *Session) getSetting() gosmpp.Settings {
    return gosmpp.Settings{
        EnquireLink:    s.EnquireLink,
        ReadTimeout:    s.ReadTimeout,
        OnAllPDU: handlePDU(s),
        OnSubmitError: s.handleSubmitError,
        OnReceivingError: s.handleReceivingError,
        OnRebindingError: s.handleRebindingError,
        OnClosed: s.handleClosed,
    }
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
        case *pdu.BindResp:
            // logrus.Info("BindResp Received")
        case *pdu.Unbind:
            fmt.Println("Unbind Received")
            return pd.GetResponse(), true
        case *pdu.UnbindResp:
            // fmt.Println("UnbindResp Received")
        case *pdu.SubmitSMResp:
            return s.handleSubmitSMResp(pd)
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
    close(s.OutstandingCh)
    close(s.stop)
    if s.transceiver != nil {
        s.transceiver.Close()
    }
    s.wg.Wait()
    // todo: is this necessary? and best place for this line ?
    close(s.ResendChannel)
    s.responseWriter.Close()
}