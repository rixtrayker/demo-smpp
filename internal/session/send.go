package session

import (
	"errors"
	"strconv"

	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/rixtrayker/demo-smpp/internal/queue"
	"github.com/sirupsen/logrus"
)

// ported redirect: zain -> mobily -> stc
var gateways = []string{"zain", "mobily", "stc"}

func (s *Session) Send(msg queue.MessageData) {
	submitSM := newSubmitSM(msg.Sender, msg.Number, msg.Text)
	logrus.Infof("Sending message to %s", msg.Number)
	ref := submitSM.SequenceNumber

	gh := append(msg.GatewayHistory, s.gateway)
	s.mu.Lock()
	s.messagesStatus[ref] = &MessageStatus{
		SystemMessageID: msg.Id,
		Sender:          msg.Sender,
		Text:            msg.Text,
		Number:          msg.Number,
		GatewayHistory:  gh,
	}
	s.mu.Unlock()

	if s.hasOutstanding {
		s.outstandingCh <- struct{}{}
	}

	err := s.send(submitSM)
	if err != nil {
		logrus.WithError(err).Error("Error sending message")
	}
}

func (s *Session) send(submitSM *pdu.SubmitSM) error {
	s.wg.Add(1)
	defer s.wg.Done()

	if s.smppSessions.transceiver != nil {
		return s.smppSessions.transceiver.Transceiver().Submit(submitSM)
	}
	if s.smppSessions.transmitter != nil {
		return s.smppSessions.transmitter.Transmitter().Submit(submitSM)
	}
	return errors.New("no valid SMPP session found")
}

// SendStream function takes a channel of MessageData and sends the messages using a limited number of goroutines
func (s *Session) SendStream(messages <-chan queue.MessageData) {
	sem := make(chan struct{}, 100) // semaphore with a capacity of 100

	for msg := range messages {
		sem <- struct{}{} // acquire a slot

		go func(m queue.MessageData) {
			defer func() { <-sem }() // release the slot
			s.Send(m)
		}(msg)
	}

	// Wait for all goroutines to finish
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
	close(s.shutdown)
}

func (s *Session) handleSubmitSMResp(pd *pdu.SubmitSMResp) {
	select {
	case <-s.outstandingCh:
		s.processSubmitSMResp(pd)
	default:
		s.processSubmitSMResp(pd)
	}
	s.wg.Done()
}

func (s *Session) processSubmitSMResp(pd *pdu.SubmitSMResp) {
	ref := pd.SequenceNumber
	s.mu.Lock()
	messageStatus := s.messagesStatus[ref]
	s.mu.Unlock()
	errCode := pd.CommandStatus.String()
	id := strconv.Itoa(messageStatus.SystemMessageID)

	s.Write(&dtos.ReceiveLog{
		MessageID:    id,
		Gateway:      s.gateway,
		MobileNo:     messageStatus.Number,
		MessageState: "Sent",
		ErrorCode:    errCode,
	})

	if pd.CommandStatus == 0 {
		logrus.Info("SubmitSMResp Received")
	} else {
		if pd.CommandStatus == data.ESME_RINVDSTADR || s.gateway == "stc" {
			go s.portMessage(messageStatus)
		}
	}
}

func (s *Session) StreamPorted() (chan queue.MessageData, chan error) {
	errors := make(chan error)
	return s.resendChannel, errors
}

func (s *Session) ClosePorted() {
	close(s.resendChannel)
}

func (s *Session) portMessage(messageStatus *MessageStatus) {
	s.wg.Add(1)
	defer s.wg.Done()
	gateway, err := s.portGateway(messageStatus.GatewayHistory)
	if err != nil {
		s.Write(&dtos.ReceiveLog{
			MessageID:    strconv.Itoa(messageStatus.SystemMessageID),
			Gateway:      s.gateway,
			MobileNo:     messageStatus.Number,
			MessageState: "Failed",
			ErrorCode:    "Unable to port message",
		})

		logrus.Error(err)
		return
	}

	s.resendChannel <- queue.MessageData{
		Id:             messageStatus.SystemMessageID,
		Gateway:        gateway,
		Sender:         messageStatus.Sender,
		Text:           messageStatus.Text,
		Number:         messageStatus.Number,
		GatewayHistory: messageStatus.GatewayHistory,
	}
}

func (s *Session) portGateway(history []string) (string, error) {
	switch len(history) {
	case 0:
		return "", errors.New("gateway history is empty, unable to port message")
	case 1:
		if s.gateway != history[0] {
			return "mobily", nil
		}
	case 2:
		if s.gateway != history[0] && s.gateway != history[1] {
			return "stc", nil
		}
	}
	return "", errors.New("unable to port message, tried all gateways")
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
