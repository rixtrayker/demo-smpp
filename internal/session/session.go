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
	transceiver    *gosmpp.Session
	ctx            context.Context
	outstandingCh  chan struct{}
	maxOutstanding int
	mu             sync.Mutex
	handler        func(pdu.PDU) (pdu.PDU, bool)
}

func CreateSessionWithRetry(ctx context.Context, provider config.Provider) (*Session, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			session, err := NewSession(ctx, provider)
			if err == nil {
				return session, nil
			}
			log.Println("Failed to create session for provider", provider.Name, err)
			log.Println("Retrying connection...")
			time.Sleep(3 * time.Second)
		}
	}
}

func NewSession(ctx context.Context, cfg config.Provider, handler func(pdu.PDU) (pdu.PDU, bool)) (*Session, error) {
	auth := gosmpp.Auth{
		SMSC:       cfg.SMSC,
		SystemID:   cfg.SystemID,
		Password:   cfg.Password,
		SystemType: "",
	}

	session := &Session{
		ctx:            ctx,
		maxOutstanding: cfg.MaxOutStanding,
		outstandingCh:  make(chan struct{}, cfg.MaxOutStanding),
	}

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

			OnAllPDU: handlePDU(session),

			OnClosed: func(state gosmpp.State) {
				fmt.Println(state)
			},
		}, 5*time.Second)

	if err != nil {
		fmt.Println("Failed to create session:", err)
		return nil, err
	}

	session.transceiver = trans
	return session, nil
}

func (s *Session) Send(message string) error {
	for {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case s.outstandingCh <- struct{}{}:
			submitSM := newSubmitSM(message)
			if err := s.transceiver.Transceiver().Submit(submitSM); err != nil {
				<-s.outstandingCh
				return err
			}
			go func() {
				defer func() { <-s.outstandingCh }()
			}()
			return nil
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
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

		case *pdu.GenericNack:
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
			defer func() { <-s.outstandingCh }()
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
	close(s.outstandingCh)
	s.transceiver.Close()
}