package smpp

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rixtrayker/demo-smpp/internal/config"
	"github.com/rixtrayker/demo-smpp/internal/session"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
)

type Session struct {
    transceiver *gosmpp.TransceiverSession
    mu          sync.Mutex
}

func NewSession(cfg config.Config, provider string) *Session {
    auth := gosmpp.Auth{
        SMSC:       cfg.SMPPConfig.SMSC,
        SystemID:   cfg.SMPPConfig.SystemID,
        Password:   cfg.SMPPConfig.Password,
        SystemType: cfg.SMPPConfig.SystemType,
    }

    trans, err := gosmpp.NewSession(
        gosmpp.TRXConnector(gosmpp.NonTLSDialer, auth),
        gosmpp.Settings{
            EnquireLink: 5 * time.Second,
            ReadTimeout: 10 * time.Second,

            OnSubmitError: func(_ pdu.PDU, err error) {
                log.Fatal("SubmitPDU error:", err)
            },

            OnReceivingError: func(err error) {
                fmt.Println("Receiving PDU/Network error:", err)
            },

            OnRebindingError: func(err error) {
                fmt.Println("Rebinding but error:", err)
            },

            OnAllPDU: handlePDU(),

            OnClosed: func(state gosmpp.State) {
                fmt.Println(state)
            },
        }, 5*time.Second)
    if err != nil {
        log.Fatal(err)
    }

    return &Session{transceiver: trans}
}

func (s *Session) Send(message string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    submitSM := newSubmitSM(message)
    if err := s.transceiver.Transceiver().Submit(submitSM); err != nil {
        fmt.Println(err)
        return err
    }
    return nil
}

func handlePDU() func(pdu.PDU) (pdu.PDU, bool) {
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
            fmt.Println("DataSM receiver")
            return pd.GetResponse(), false

        case *pdu.DeliverSM:
            fmt.Println("DeliverSM receiver")
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