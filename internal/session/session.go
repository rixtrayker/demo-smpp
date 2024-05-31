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
	transceiver   *gosmpp.Session
	maxOutstanding int
	outstandingCh chan struct{}
	mu            sync.Mutex
	handler       func(pdu.PDU) (pdu.PDU, bool)
	stop chan struct{}
	concatenated map[uint8][]string
	Status map[int32]*MessageStatus
	maxRetries    int
	responseWriter *response.Writer
	wg 				sync.WaitGroup
	notCancelled bool
}

type MessageStatus struct {
	DreamsMessageID int
	MessageID string
	Status string
	Number string
}

func NewSession(cfg config.Provider, handler func(pdu.PDU) (pdu.PDU, bool),rw *response.Writer) *Session {
	session := &Session{
		concatenated: make(map[uint8][]string),
		handler:      handler,
		maxRetries:     cfg.MaxRetries,
		maxOutstanding: cfg.MaxOutStanding,
		outstandingCh:  make(chan struct{}, cfg.MaxOutStanding),
		stop: make(chan struct{}),
		Status: make(map[int32]*MessageStatus),
		responseWriter:  rw,
		wg: sync.WaitGroup{},
		notCancelled: true,
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
			s.Status[ref].MessageID = msgID
			s.Status[ref].Status = "pending"
			s.mu.Unlock()
            s.responseWriter.WriteResponse(&dtos.ReceiveLog{
                MessageID: msgID,
                MobileNo: s.Status[ref].Number,
                MessageState: "SubmitSMResp",
            })
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
    for; len(s.outstandingCh) > 0; {}
	close(s.stop)
	if s.transceiver != nil {
		s.transceiver.Close()
	}
	s.wg.Wait()
    s.responseWriter.Close()
}