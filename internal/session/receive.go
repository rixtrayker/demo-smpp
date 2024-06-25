package session

import (
	"errors"
	"strings"

	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/db"
	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/rixtrayker/demo-smpp/internal/metrics"
	"github.com/rixtrayker/demo-smpp/internal/models"
	"github.com/sirupsen/logrus"
)

func (s *Session) HandleDeliverSM(pd *pdu.DeliverSM) {
	defer s.deliveryWg.Done()
	msg, err := pd.Message.GetMessage()
	status := ""
	if err != nil {
		logError(err, pd)
		return
	}

	errCode := extractField(msg, "err:")
	if errCode == "021" && strings.EqualFold(s.gateway, "STC") {
		if err := s.deliverPort(pd); err != nil {
			logrus.WithError(err).Info("Error in deliverPort")
			status = "Porting Failed"
		}
	}

	data := s.prepareResult(pd, status)
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

func (s *Session) prepareResult(pd *pdu.DeliverSM, status string) dtos.ReceiveLog {
	msgID := getReceiptedMessageID(pd)
	mobileNo := pd.SourceAddr.Address()
	submitID, _, errCode, err := getMessageData(pd)
	if err != nil {
		logrus.WithError(err).Info("Got error when getting DeliverSM message")
	}
	if status == "" {
		if errCode == "000" {
			status = "Delivered"
		} else {
			status = "Undelivered"
		}	
	}

	return dtos.ReceiveLog{
		// id will be added periodically
		// join the DeliverSM with its SubmitSMResp on message_id
		// SystemMessageID: 
		Gateway:	  s.gateway,
		MessageID:    msgID,
		MobileNo:     mobileNo,
		MessageState: status,
		ErrorCode:    errCode,
		Data:         "id:" + submitID,
	}
}

// stc on error code 21
func (s *Session) deliverPort(pd *pdu.DeliverSM) error {
	msgID := getReceiptedMessageID(pd)

	var dlrSms []models.DlrSms 
	result := db.DB.Where("messageId = ?", msgID).Where("mobileNo = ?", pd.DestAddr.Address()).Where("messageState = ?", "Sent").Find(&dlrSms)
	
	// if dlrSms len is 3 so we tried all gateway and don't port return an error
	if len(dlrSms) == 3 {
		return errors.New("You tried 3 gateways on msgID:" + msgID)
	}

	// history := []string{}

	// for _, dlr := range dlrSms {
	// 	history = append(history, dlr.Gateway)
	// }

	history := []string{"stc"}

	if result.Error != nil {
		metrics.PortedMessages.WithLabelValues("db-error", "stc", "zain").Inc()
		logrus.WithError(result.Error).Info("Error getting DlrSms")
		return result.Error
	}

	msg := &MessageStatus{
		SystemMessageID: dlrSms[0].ID,
		Sender:          pd.SourceAddr.Address(),
		Number:          pd.DestAddr.Address(),
		Text:            dlrSms[0].Data,
		MessageID:       msgID,
		GatewayHistory:  history,
	}

	s.portMessage(msg)
	return nil
}

func getReceiptedMessageID(pd *pdu.DeliverSM) string {
	tag := pdu.TagReceiptedMessageID
	pduField := pd.OptionalParameters[tag]
	return pduField.String()
}
