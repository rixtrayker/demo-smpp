package session

import (
	"log"
	"strings"

	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/sirupsen/logrus"
)


func (s *Session) handleDeliverSM(pd *pdu.DeliverSM) (pdu.PDU, bool) {
	// receiver := pd.DestAddr.Address()
	message, err := pd.Message.GetMessage()
	if err != nil {
		logrus.WithError(err).Info("Decoding DeliverSM message error")
	}

	totalParts, sequence, reference, found := pd.Message.UDH().GetConcatInfo()
	// udh:= pd.Message.UDH()
	// log.Printf("udh: %v", udh)
	
	data := PrepareResult(pd)
	(*s.responseWriter).WriteResponse(&data)

	if found {
		return s.handleConcatenatedSMS(reference, message, totalParts, sequence, pd)
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

func getRceiptedMessageID(pd *pdu.DeliverSM) string {
	tag := pdu.TagReceiptedMessageID
	pduField := pd.OptionalParameters[tag]
	msgId := pduField.String()
	return msgId
}

func getMessageId(pd *pdu.DeliverSM) string {
	msgID := ""
	message, err := pd.Message.GetMessage()
	if err != nil {
		logrus.WithError(err).Info("Got error when getting DeliverSM message")
	}
	result := strings.Split(message, "msgID:")

	if len(result) > 1 {
		msgID = result[1]
	}

	result = strings.Split(msgID, " ")

	if len(result) > 1 {
		msgID = result[0]
	}

	return msgID
}

func (s *Session) Write(rl dtos.ReceiveLog){
	(*s.responseWriter).WriteResponse(&rl)
}

func PrepareResult(pd *pdu.DeliverSM) dtos.ReceiveLog {
	msgID := getRceiptedMessageID(pd)
	mobileNo := pd.SourceAddr.Address()

		// MessageState string 
		// ErrorCode    string 
		// MobileNo     int64  
		// CurrentTime  time.Time
		// Data         string 

	return dtos.ReceiveLog{
		MessageID: msgID,
		MobileNo: mobileNo,
	}   
}
