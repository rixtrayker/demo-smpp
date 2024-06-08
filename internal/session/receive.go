package session

import (
	"strings"

	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/db"
	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/rixtrayker/demo-smpp/internal/models"
	"github.com/sirupsen/logrus"
)

func (s *Session) HandleDeliverSM(pd *pdu.DeliverSM) {
	msg, err := pd.Message.GetMessage()
	if err != nil {
		logError(err, pd)
		return
	}

	errCode := extractField(msg, "err:")
	if errCode == "021" && strings.EqualFold(s.gateway, "STC") {
		if err := s.deliverPort(pd); err != nil {
			logrus.WithError(err).Info("Error in deliverPort")
		}
	}

	data := s.prepareResult(pd)
	s.Write(&data)
}

func logError(err error, pd *pdu.DeliverSM) {
	logrus.WithError(err).Info("Error getting DeliverSM message")
	logrus.WithFields(logrus.Fields{
		"source": pd.SourceAddr.Address(),
		"dest":   pd.DestAddr.Address(),
	}).Info("Error getting DeliverSM message")
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

func getMessageData(pd *pdu.DeliverSM) (string, string, string, error) {
	message, err := pd.Message.GetMessage()
	if err != nil {
		logrus.WithError(err).Info("Error getting DeliverSM message")
		return "", "", "", err
	}

	id := extractField(message, "id:")
	dlvrd := extractField(message, "dlvrd:")
	errCode := extractField(message, "err:")

	return id, dlvrd, errCode, nil
}

func (s *Session) Write(rl *dtos.ReceiveLog) {
	(*s.responseWriter).WriteResponse(rl)
}

func (s *Session) prepareResult(pd *pdu.DeliverSM) dtos.ReceiveLog {
	msgID := getReceiptedMessageID(pd)
	mobileNo := pd.SourceAddr.Address()
	submitID, dlvrd, errCode, err := getMessageData(pd)
	if err != nil {
		logrus.WithError(err).Info("Got error when getting DeliverSM message")
	}

	status := "Undelivered"
	if dlvrd == "000" {
		status = "Delivered"
	}

	return dtos.ReceiveLog{
		MessageID:    msgID,
		MobileNo:     mobileNo,
		MessageState: status,
		ErrorCode:    errCode,
		Data:         "id:" + submitID,
	}
}

func (s *Session) deliverPort(pd *pdu.DeliverSM) error {
	msgID := getReceiptedMessageID(pd)

	var dlrSms models.DlrSms
	result := db.DB.Select("*").Where("messageId = ?", msgID).First(&dlrSms)
	if result.Error != nil {
		logrus.WithError(result.Error).Info("Error getting DlrSms")
		return result.Error
	}

	msg := &MessageStatus{
		SystemMessageID: int(dlrSms.ID),
		Sender:          pd.SourceAddr.Address(),
		Number:          pd.DestAddr.Address(),
		Text:            dlrSms.Data,
		MessageID:       msgID,
		GatewayHistory:  []string{s.gateway},
	}

	s.portMessage(msg)
	return nil
}

func getReceiptedMessageID(pd *pdu.DeliverSM) string {
	tag := pdu.TagReceiptedMessageID
	pduField := pd.OptionalParameters[tag]
	return pduField.String()
}
