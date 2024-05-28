package session

import (
	"context"
	"fmt"
	"log"
	"strings"
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

// may need to pass WG to wait for all sessions to close
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
            log.Printf("Failed to create session for provider: %v, error: %v\n", cfg.Name, err)

            if strings.Contains(err.Error(), "connect: connection refused") {
                log.Printf("Connection refused. Ensure the SMSC server is running and accessible on %s\n", cfg.SMSC)
            }

            time.Sleep(delay)
        }
    }
	return fmt.Errorf("failed to create session after %d retries", s.maxRetries)
}

func (s *Session) Send(sender, number, message string) error {
    submitSM := newSubmitSM(sender, number, message)
	ref := submitSM.SequenceNumber
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status[ref] = &MessageStatus{
		// DreamsMessageID: dreamsMessageId,
		Number: number,
	}

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

func newSubmitSM(sender, number, message string) *pdu.SubmitSM {
    srcAddr := pdu.NewAddress()
    srcAddr.SetTon(5)
    srcAddr.SetNpi(0)
    _ = srcAddr.SetAddress(sender)

    destAddr := pdu.NewAddress()
    destAddr.SetTon(1)
    destAddr.SetNpi(1)
    _ = destAddr.SetAddress(number)

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

func (s *Session) handleDeliverSM(pd *pdu.DeliverSM) (pdu.PDU, bool) {
	message, err := pd.Message.GetMessage()
	if err != nil {
		log.Fatal(err)
	}

	// receiver := pd.DestAddr.Address()


	totalParts, sequence, reference, found := pd.Message.UDH().GetConcatInfo()
	// udh:= pd.Message.UDH()
	// log.Printf("udh: %v", udh)

	if found {
		return s.handleConcatenatedSMS(reference, message, totalParts, sequence, pd)
	}
	msgID := ""
	result := strings.Split(message, "msgID:")
	if len(result) > 1 {
		msgID = result[1]
	}
	result = strings.Split(msgID, " ")
	if len(result) > 1 {
		msgID = result[0]
		ref := pd.SequenceNumber
		s.mu.Lock()
		defer s.mu.Unlock()
		if _, ok := s.Status[ref]; ok {
			if s.Status[ref].MessageID == msgID {
				s.Status[ref].Status = "sent"
			} else {
				// push deliver status to queue to 
				
			}
		}
	}
	return pd.GetResponse(), false
}

func (s *Session) handleConcatenatedSMS(reference uint8, message string, totalParts, sequence byte, pd *pdu.DeliverSM) (pdu.PDU, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.concatenated[reference]; !ok {
		s.concatenated[reference] = make([]string, totalParts)
	}

	s.concatenated[reference][sequence-1] = message

	if isConcatenatedDone(s.concatenated[reference], totalParts) {
		log.Println(strings.Join(s.concatenated[reference], ""))
		delete(s.concatenated, reference)
	}

	return pd.GetResponse(), false
}

func isConcatenatedDone(parts []string, total byte) bool {
	for _, part := range parts {
		if part == "" {
			total--
		}
	}
	return total == 0
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