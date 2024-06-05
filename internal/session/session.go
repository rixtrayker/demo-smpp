package session

import (
	"errors"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/queue"
	"github.com/rixtrayker/demo-smpp/internal/response"
	"github.com/sirupsen/logrus"
)

type SessionType string

const (
	Transceiver SessionType = "transceiver"
	Receiver    SessionType = "receiver"
	Transmitter SessionType = "transmitter"
)

type Session struct {
	gateway            string
	sessionType        SessionType
	maxOutstanding     int
	hasOutstanding     bool
	outstandingCh      chan struct{}
	resendChannel      chan queue.MessageData
	mu                 sync.Mutex
	handler            *PDUHandler
	stop               chan struct{}
	concatenated       map[uint8][]string
	messagesStatus     map[int32]*MessageStatus
	maxRetries         int
	responseWriter     *response.Writer
	wg                 sync.WaitGroup
	auth               gosmpp.Auth
	enquireLink        time.Duration
	readTimeout        time.Duration
	rebindingInterval  time.Duration
	portGateways       []string
	smppSessions       SMPPSessions
}

type MessageStatus struct {
	MessageID       string
	SystemMessageID int
	Sender          string
	Text            string
	Status          string
	Number          string
	GatewayHistory  []string
}

type SMPPSessions struct {
	transceiver *gosmpp.Session
	receiver    *gosmpp.Session
	transmitter *gosmpp.Session
}

type PDUHandler interface {
	HandlePDU(*interface{}) (pdu.PDU, bool)
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

func NewSession(cfg config.Provider, h *PDUHandler, options ...Option) (*Session, error) {
	session := &Session{
		gateway:         cfg.Name,
		concatenated:    make(map[uint8][]string),
		handler:         h,
		maxOutstanding:  cfg.MaxOutStanding,
		hasOutstanding:  cfg.HasOutStanding,
		maxRetries:      cfg.MaxRetries,
		outstandingCh:   make(chan struct{}, cfg.MaxOutStanding),
		stop:            make(chan struct{}),
		messagesStatus:  make(map[int32]*MessageStatus),
		wg:              sync.WaitGroup{},
		sessionType:     SessionType(cfg.SessionType),
		enquireLink:     5 * time.Second,
		readTimeout:     10 * time.Second,
		rebindingInterval: 600 * time.Second,
		resendChannel:   make(chan queue.MessageData),
		portGateways:    []string{"zain", "mobily", "stc"},
		smppSessions:    SMPPSessions{},
	}

	for _, option := range options {
		option(session)
	}

	session.auth = gosmpp.Auth{
		SMSC:       cfg.SMSC,
		SystemID:   cfg.SystemID,
		Password:   cfg.Password,
		SystemType: cfg.SystemType,
	}

	if err := session.connectSessions(); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Session) Start() error {
	initialDelay := 100 * time.Millisecond
	maxDelay := 10 * time.Second
	factor := 2.0

	for retries := 0; retries <= s.maxRetries; retries++ {
		select {
		case <-s.stop:
			return errors.New("session stopped")
		default:
			err := s.connectSessions()
			if err != nil {
				delay := calculateBackoff(initialDelay, maxDelay, factor, retries)
				logrus.WithError(err).Errorf("Failed to create session for provider %s", s.gateway)
				time.Sleep(delay)
			} else {
				return nil
			}
		}
	}

	return errors.New("failed to create session after maximum retries")
}

func (s *Session) connectSessions() error {
	var err error
	switch s.sessionType {
	case Transceiver:
		s.smppSessions.transceiver, err = gosmpp.NewSession(
			gosmpp.TRXConnector(gosmpp.NonTLSDialer, s.auth),
			s.getSettings(),
			s.rebindingInterval,
		)
		if err != nil {
			return err
		}
	default:
		s.smppSessions.receiver, err = gosmpp.NewSession(
			gosmpp.RXConnector(gosmpp.NonTLSDialer, s.auth),
			s.getSettings(),
			s.rebindingInterval,
		)
		if err != nil {
			return err
		}

		s.smppSessions.transmitter, err = gosmpp.NewSession(
			gosmpp.TXConnector(gosmpp.NonTLSDialer, s.auth),
			s.getSettings(),
			s.rebindingInterval,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Session) getSettings() gosmpp.Settings {
	return gosmpp.Settings{
		EnquireLink:       s.enquireLink,
		ReadTimeout:       s.readTimeout,
		OnAllPDU:          handlePDU(s),//s.handler.HandlePDU,
		OnSubmitError:     s.handleSubmitError,
		OnReceivingError:  s.handleReceivingError,
		OnRebindingError:  s.handleRebindingError,
		OnClosed:          s.handleClosed,
	}
}

func (s *Session) handleSubmitError(p pdu.PDU, err error) {
	logrus.WithError(err).Error("SubmitPDU error")
}

func (s *Session) handleReceivingError(err error) {
	logrus.WithError(err).Error("Receiving PDU/Network error")
}

func (s *Session) handleRebindingError(err error) {
	logrus.WithError(err).Error("Rebinding error")
}

func (s *Session) handleClosed(state gosmpp.State) {
    logrus.Info("Session closed: ", state)
}


func handlePDU2(pd pdu.PDU) (pdu.PDU, bool) {
    return pd.GetResponse(), true
}

func handlePDU(s *Session) func(pdu.PDU) (pdu.PDU, bool) {
    return func(p pdu.PDU) (pdu.PDU, bool) {
        switch pd := p.(type) {
        case *pdu.BindResp:
            // logrus.Info("BindResp Received")
        case *pdu.Unbind:
            logrus.Info("Unbind Received")
            return pd.GetResponse(), true
        case *pdu.UnbindResp:
            // fmt.Println("UnbindResp Received")
        case *pdu.SubmitSMResp:
            s.handleSubmitSMResp(pd)
            return pd.GetResponse(), false
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
            s.HandleDeliverSM(pd)
            return pd.GetResponse(), false
        }
        return nil, false
    }
}

func (s *Session) Stop() {
    close(s.outstandingCh)
    close(s.stop)

    if s.smppSessions.transceiver != nil {
        s.smppSessions.transceiver.Close()
    }
    s.wg.Wait()

    close(s.resendChannel)
    (*s.responseWriter).Close()
}

type Handler struct {
}

func (h *Handler) HandlePDU(pd *interface{}) (pdu.PDU, bool) {
    return (*pd).(pdu.PDU).GetResponse(), true
}
