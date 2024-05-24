package session

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/config"
)

type Session struct {
    transceiver   *gosmpp.Session
    ctx           context.Context
    maxOutstanding int
    mu            sync.Mutex
    handler       func(pdu.PDU) (pdu.PDU, bool)
    maxRetries    int
    outstandingCh chan struct{}
}

func NewSession(ctx context.Context, cfg config.Provider, handler func(pdu.PDU) (pdu.PDU, bool)) (*Session, error) {
	session := &Session{
		ctx:            ctx,
		maxOutstanding: cfg.MaxOutStanding,
		handler:        handler,
		maxRetries:     cfg.MaxRetries,
        outstandingCh:  make(chan struct{}, cfg.MaxOutStanding),
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
			}, 5*time.Second)

		if err == nil {
			s.transceiver = trans
			return nil
		}

		delay := calculateBackoff(initialDelay, maxDelay, factor, retries)
		log.Printf("Failed to create session, retrying in %v: %v", delay, err)
		time.Sleep(delay)
	}

	return fmt.Errorf("failed to create session after %d retries", s.maxRetries)
}

func (s *Session) Send(message string) error {
    submitSM := newSubmitSM(message)
    s.outstandingCh <- struct{}{}
    err := s.transceiver.Transceiver().Submit(submitSM)
    fmt.Println("SubmitSM: " + message)
    if err != nil {
        log.Println("SubmitPDU error:", err)
    }

    return nil
}

func handlePDU(s *Session) func(pdu.PDU) (pdu.PDU, bool) {
    return func(p pdu.PDU) (pdu.PDU, bool) {
        switch pd := p.(type) {
        case *pdu.Unbind:
            fmt.Println("Unbind Received")
            return pd.GetResponse(), true
        case *pdu.UnbindResp:
            fmt.Println("UnbindResp Received")
        case *pdu.SubmitSMResp:
            fmt.Println("SubmitSMResp Received")
            <-s.outstandingCh
        case *pdu.GenericNack:
            <-s.outstandingCh
            fmt.Println("GenericNack Received")
        case *pdu.EnquireLinkResp:
            fmt.Println("EnquireLinkResp Received")
        case *pdu.EnquireLink:
            fmt.Println("EnquireLink Received")
            return pd.GetResponse(), false
        case *pdu.DataSM:
            fmt.Println("DataSM Received")
            return pd.GetResponse(), false
        case *pdu.DeliverSM:
            fmt.Println("DeliverSM Received")
            return pd.GetResponse(), false
        }
        return nil, false
    }
}

func newSubmitSM(message string) *pdu.SubmitSM {
    srcAddr := pdu.NewAddress()
    srcAddr.SetTon(5)
    srcAddr.SetNpi(0)
    _ = srcAddr.SetAddress("00" + "522241")

    destAddr := pdu.NewAddress()
    destAddr.SetTon(1)
    destAddr.SetNpi(1)
    _ = destAddr.SetAddress("99" + "522241")

    submitSM := pdu.NewSubmitSM().(*pdu.SubmitSM)
    submitSM.SourceAddr = srcAddr
    submitSM.DestAddr = destAddr
    _ = submitSM.Message.SetMessageWithEncoding(message, data.UCS2)
    submitSM.ProtocolID = 0
    submitSM.RegisteredDelivery = 1
    submitSM.ReplaceIfPresentFlag = 0
    submitSM.EsmClass = 0

    return submitSM
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