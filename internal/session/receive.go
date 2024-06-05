package session

import (
	"strings"

	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/sirupsen/logrus"
)

func (s *Session) HandleDeliverSM(pd *pdu.DeliverSM) {
    data := prepareResult(pd)
    s.Write(&data)
}

func getReceiptedMessageID(pd *pdu.DeliverSM) string {
	tag := pdu.TagReceiptedMessageID
	pduField := pd.OptionalParameters[tag]
	return pduField.String()
}

func extractField(message, prefix string) string {
	parts := strings.SplitN(message, prefix, 2)
	if len(parts) < 2 {
		return ""
	}
	field := strings.Fields(parts[1])
	if len(field) > 0 {
		return field[0]
	}
	return ""
}

func getMessageData(pd *pdu.DeliverSM) (string, string, error) {
	message, err := pd.Message.GetMessage()
	if err != nil {
		logrus.WithError(err).Info("Error getting DeliverSM message")
		return "", "", err
	}

	id := extractField(message, "id:")
	dlvrd := extractField(message, "dlvrd:")

	return id, dlvrd, nil
}

// pass by reference later, if it is better
func (s *Session) Write(rl *dtos.ReceiveLog) {
	(*s.responseWriter).WriteResponse(rl)
}

func prepareResult(pd *pdu.DeliverSM) dtos.ReceiveLog {
	msgID := getReceiptedMessageID(pd)
	mobileNo := pd.SourceAddr.Address()
	submitID, dlvrd, err := getMessageData(pd)
	if err != nil {
        logrus.WithError(err).Info("Got error when getting DeliverSM message")
	}

	return dtos.ReceiveLog{
        MessageID:    msgID,
        MobileNo:     mobileNo,
        MessageState: "DELIVERED",
        ErrorCode:    dlvrd,
        Data:         "id:" + submitID,
    }
}
