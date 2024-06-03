package session

import (
	"strings"

	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/sirupsen/logrus"
)


func (s *Session) handleDeliverSM(pd *pdu.DeliverSM) (pdu.PDU, bool) {
	data := PrepareResult(pd)
	(*s.responseWriter).WriteResponse(&data)

	return pd.GetResponse(), false
}

func getRceiptedMessageID(pd *pdu.DeliverSM) string {
	tag := pdu.TagReceiptedMessageID
	pduField := pd.OptionalParameters[tag]
	msgId := pduField.String()
	return msgId
}

func getField(message string, prefix string) string {
	field := ""

	result := strings.Split(message, prefix)

	if len(result) > 1 {
		field = result[1]
	}

	result = strings.Split(field, " ")

	if len(result) > 1 {
		field = result[0]
	}

	return field
}


func getMessageData(pd *pdu.DeliverSM) (string, string, error) {
	message, err := pd.Message.GetMessage()
	if err != nil {
		logrus.WithError(err).Info("Got error when getting DeliverSM message")
		return "", "", err
	}

	id := getField(message, "id:")
	dlvrd := getField(message, "dlvrd:")

	return id, dlvrd, nil
}

func (s *Session) Write(rl dtos.ReceiveLog){
	(*s.responseWriter).WriteResponse(&rl)
}

func PrepareResult(pd *pdu.DeliverSM) dtos.ReceiveLog {
	msgID := getRceiptedMessageID(pd)
	mobileNo := pd.SourceAddr.Address()
	submitID, dlvrd, err := getMessageData(pd)
	if err != nil {
		logrus.WithError(err).Info("Got error when getting DeliverSM message")
	}

	return dtos.ReceiveLog{
		MessageID: msgID,
		MobileNo: mobileNo,
		MessageState: "DELIVERED",
		ErrorCode: dlvrd,
		Data: "id:" + submitID,
	}   
}
