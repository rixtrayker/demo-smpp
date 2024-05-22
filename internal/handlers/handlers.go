package handlers

import (
	"fmt"
	"log"

	"github.com/linxGnu/gosmpp/pdu"
)

func ProviderAHandler(p pdu.PDU) (pdu.PDU, bool) {
	// Handler logic for provider A
	return handlePDU(p)
}

func ProviderBHandler(p pdu.PDU) (pdu.PDU, bool) {
	// Handler logic for provider B
	return handlePDU(p)
}

func ProviderCHandler(p pdu.PDU) (pdu.PDU, bool) {
	// Handler logic for provider C
	return handlePDU(p)
}

func handlePDU(p pdu.PDU) (pdu.PDU, bool) {
	concatenated := map[uint8][]string{}
	switch pd := p.(type) {
	case *pdu.SubmitSMResp:
		fmt.Println("SubmitSMResp:")//pd)

	case *pdu.GenericNack:
		// fmt.Println("GenericNack Received")

	case *pdu.EnquireLinkResp:
		// fmt.Println("EnquireLinkResp Received")

	case *pdu.DataSM:
		// fmt.Printf("DataSM:")//pd)

	case *pdu.DeliverSM:
		// fmt.Printf("DeliverSM:")//pd)
		// log.Println(pd.Message.GetMessage())
		// region concatenated sms (sample code)
		message, err := pd.Message.GetMessage()
		if err != nil {
			log.Fatal(err)
		}
		totalParts, sequence, reference, found := pd.Message.UDH().GetConcatInfo()
		if found {
			if _, ok := concatenated[reference]; !ok {
				concatenated[reference] = make([]string, totalParts)
			}
			concatenated[reference][sequence-1] = message
		}
		if !found {
			// log.Println(message)
		} else if _, ok := concatenated[reference]; ok {
			log.Println(concatenated[reference])
			// delete(concatenated, reference)
		}
		// endregion
	}
	return nil, false
}