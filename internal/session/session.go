package session

import (
	"errors"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/dtos"
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
	gateway           string
	sessionType       SessionType
	maxOutstanding    int
	hasOutstanding    bool
	outstandingCh     chan struct{}
	resendChannel     chan queue.MessageData
	mu                sync.Mutex
	handler           *PDUHandler
	stop              chan struct{}
	concatenated      map[uint8][]string
	messagesStatus    map[int32]*MessageStatus
	maxRetries        int
	responseWriter    *response.Writer
	wg                sync.WaitGroup
	deliveryWg        sync.WaitGroup
	auth              gosmpp.Auth
	enquireLink       time.Duration
	readTimeout       time.Duration
	rebindingInterval time.Duration
	portGateways      []string
	smppSessions      SMPPSessions
	shutdown 		  CloseSignals
}

type CloseSignals struct {
	streamClose chan struct{}
	portedClosed chan struct{}
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
		gateway:           cfg.Name,
		concatenated:      make(map[uint8][]string),
		handler:           h,
		maxOutstanding:    cfg.MaxOutStanding,
		hasOutstanding:    cfg.HasOutStanding,
		maxRetries:        cfg.MaxRetries,
		outstandingCh:     make(chan struct{}, cfg.MaxOutStanding),
		stop:              make(chan struct{}),
		messagesStatus:    make(map[int32]*MessageStatus),
		wg:                sync.WaitGroup{},
		deliveryWg:		   sync.WaitGroup{},
		sessionType:       SessionType(cfg.SessionType),
		enquireLink:       5 * time.Second,
		readTimeout:       10 * time.Second,
		rebindingInterval: 600 * time.Second,
		resendChannel:     make(chan queue.MessageData),
		portGateways:      []string{"zain", "mobily", "stc"},
		smppSessions:      SMPPSessions{},
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
			if err := s.connectSessions(); err != nil {
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
		EnquireLink:      s.enquireLink,
		ReadTimeout:      s.readTimeout,
		OnAllPDU:         handlePDU(s),
		OnSubmitError:    s.handleSubmitError,
		OnReceivingError: s.handleReceivingError,
		OnRebindingError: s.handleRebindingError,
		OnClosed:         s.handleClosed,
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

func handlePDU(s *Session) func(pdu.PDU) (pdu.PDU, bool) {
	return func(p pdu.PDU) (pdu.PDU, bool) {
		switch pd := p.(type) {
		case *pdu.BindResp:
			// Handle BindResp if needed
		case *pdu.Unbind:
			logrus.Info("Unbind Received")
			return pd.GetResponse(), true
		case *pdu.UnbindResp:
			// Handle UnbindResp if needed
		case *pdu.SubmitSMResp:
			s.handleSubmitSMResp(pd)
			return pd.GetResponse(), false
		case *pdu.GenericNack:
			// Handle GenericNack if needed
		case *pdu.EnquireLinkResp:
			// Handle EnquireLinkResp if needed
		case *pdu.EnquireLink:
			return pd.GetResponse(), false
		case *pdu.DataSM:
			return pd.GetResponse(), false
		case *pdu.DeliverSM:
			s.deliveryWg.Add(1)
			s.HandleDeliverSM(pd)
			return pd.GetResponse(), false
		}
		return nil, false
	}
}


func (s *Session) writeLog(log *dtos.ReceiveLog) {
	if s.responseWriter != nil {
		(*s.responseWriter).WriteResponse(log)
	}
}

type Handler struct {
}

func (h *Handler) HandlePDU(pd *interface{}) (pdu.PDU, bool) {
    return (*pd).(pdu.PDU).GetResponse(), true
}

func (s *Session) PushResendChannel(msg queue.MessageData) {
	s.resendChannel <- msg
}

// func (s *Session) PopResendChannel() queue.MessageData {
// 	msg := <-s.resendChannel
// 	return msg
// }

func (s *Session) ShutdownSignals() {
	<-s.shutdown.streamClose
}

func (s *Session) Stop() {
	s.ShutdownSignals()
	if s.smppSessions.transceiver != nil {
		s.smppSessions.transceiver.Close()
	}
	if s.smppSessions.receiver != nil {
		s.smppSessions.receiver.Close()
	}
	if s.smppSessions.transmitter != nil {
		s.smppSessions.transmitter.Close()
	}
	close(s.outstandingCh)
	close(s.stop) // who is lestinging to this channel?
	
    s.wg.Wait()

	<-s.shutdown.portedClosed
	s.deliveryWg.Wait()
	if s.responseWriter != nil {
		(*s.responseWriter).Close()
	}
}