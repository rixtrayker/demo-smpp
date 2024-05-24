package handlers

import (
	"fmt"

	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/session"
)

func ProviderAHandler(s *session.Session) func(p pdu.PDU) (pdu.PDU, bool) {
	// Handler logic for provider A
	return handlePDU(s)
}

func ProviderBHandler(s *session.Session) func(p pdu.PDU) (pdu.PDU, bool) {
	// Handler logic for provider B
	return handlePDU(s)
}

func ProviderCHandler(s *session.Session) func(p pdu.PDU) (pdu.PDU, bool) {
	// Handler logic for provider C
	return handlePDU(s)
}


func handlePDU(s *session.Session) func(pdu.PDU) (pdu.PDU, bool) {
	return func(p pdu.PDU) (pdu.PDU, bool) {
		switch pd := p.(type) {
		case *pdu.Unbind:
			fmt.Println("Unbind Received")
			return pd.GetResponse(), true

		case *pdu.UnbindResp:
			fmt.Println("UnbindResp Received")

		case *pdu.SubmitSMResp:
			// <-s.outstandingCh
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
			return pd.GetResponse(), false
		}
		return nil, false
	}
}