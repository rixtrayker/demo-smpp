package session

import (
	"log"
	"strings"

	"github.com/linxGnu/gosmpp/pdu"
)


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