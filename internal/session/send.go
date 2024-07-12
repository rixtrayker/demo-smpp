package session

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/rixtrayker/demo-smpp/internal/metrics"
	"github.com/rixtrayker/demo-smpp/internal/queue"
	"github.com/sirupsen/logrus"
)

func (s *Session) Send(msg queue.MessageData) error {
	submitSM := newSubmitSM(msg.Sender, msg.Number, msg.Text)
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

	return s.send(submitSM)
}

func (s *Session) send(submitSM *pdu.SubmitSM) error {
	s.allowSend()
	var err error
	
	if s.smppSessions.transceiver != nil {
		err = s.smppSessions.transceiver.Transceiver().Submit(submitSM)
	}
	if s.smppSessions.transmitter != nil {
		err = s.smppSessions.transmitter.Transmitter().Submit(submitSM)
	}

	if err == nil{
		s.wg.Add(1)	
		return nil
	} else {
		s.retrySend(s.messagesStatus[submitSM.SequenceNumber], err.Error())
		if err == gosmpp.ErrConnectionClosing {
			logrus.Error("Connection closing")
			return s.connectRetry(s.ctx)
		}
	}

	return err
}


func (s *Session) allowSend() bool {
	if err := s.limiter.Wait(context.Background()); err == nil {
		return true
	}
	return false
}

func (s *Session) SendStream(messages <-chan queue.MessageData){
	sem := make(chan struct{}, 100) // semaphore with a capacity of 100
	// errChan := make(chan error)
	for msg := range messages {
		sem <- struct{}{} // acquire a slot
		go func(m queue.MessageData) {
			defer func() { <-sem }() // release the slot
			err := s.Send(m)
			if err != nil {
				if err == gosmpp.ErrConnectionClosing {
					logrus.Error("Connection closing")
				}
				// go func(err error) {
				// 		errChan <- err
				// }(err)
			}
		}(msg)
	}

	// Wait for all goroutines to finish
	for i := 0; i < cap(sem); i++ {
		sem <- struct{}{}
	}
	close(s.shutdown.streamClose)
	// return errChan
}

func (s *Session) SendStreamWithCancel(ctx context.Context, messages <-chan queue.MessageData) {
	defer close(s.shutdown.streamClose)

	for msg := range messages {
		select {
		case <-ctx.Done():
			s.resendStream.stream <- msg
			return
		default:
			err := s.Send(msg)
			if err != nil {
				if err == gosmpp.ErrConnectionClosing {
					logrus.Error("Connection closing")
				}
			}
		}
	}

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
	
	if messageStatus == nil {
		s.responseWriter.WriteResponse(&dtos.ReceiveLog{
			SystemMessageID: 0,
			MessageID:       pd.MessageID,
			Gateway:         s.gateway,
			MobileNo:        "",
			MessageState:    "Failed",
			ErrorCode:       pd.CommandStatus.String(),
			Data:            "SubmitSMResp: Message not found",
		})
		return
	}

	st := messageStatus.startTime
	metrics.SubmitSMRespDuration.Observe(time.Since(st).Seconds())

	errCode := pd.CommandStatus.String()
	status := ""

	new_or_ported := "new"
	if len(messageStatus.GatewayHistory) > 1 {
		new_or_ported = "ported"
	}
	
	if pd.IsOk() {
		status = "Sent"
		metrics.SentMessages.WithLabelValues(status, s.gateway, new_or_ported).Inc()
	} else {
		if pd.CommandStatus == data.ESME_RINVDSTADR || s.gateway != "stc" { // ensure logic
			go s.portMessage(messageStatus)
			status = "Ported"
			// metrics.SentMessages.WithLabelValues(status, s.gateway, new_or_ported).Inc()
		} else {
			if pd.CommandStatus == data.ESME_RSUBMITFAIL || pd.CommandStatus == data.ESME_RTHROTTLED {
				s.retrySend(messageStatus, pd.CommandStatus.String())
				status = "Resend"
			} else {
				status = "Failed"
				metrics.SentMessages.WithLabelValues(status, s.gateway, new_or_ported).Inc()
			}
		}
	}

	go func() {
		s.Write(&dtos.ReceiveLog{
			SystemMessageID: messageStatus.SystemMessageID,
			MessageID:       pd.MessageID,
			Gateway:         s.gateway,
			MobileNo:        messageStatus.Number,
			MessageState:    status,
			ErrorCode:       errCode,
			Data:            messageStatus.Text,
		})
	}()
}

func (s *Session) StreamPorted() (chan queue.MessageData, chan error) {
	return s.portedStream.stream, s.portedStream.err
}

func (s *Session) retrySend(messageStatus *MessageStatus, status string) { 
	metrics.ResendMessages.WithLabelValues(status, s.gateway).Inc()

	s.resendStream.stream <- queue.MessageData{
		Id:             messageStatus.SystemMessageID,
		Gateway:        s.gateway,
		Sender:         messageStatus.Sender,
		Text:           messageStatus.Text,
		Number:         messageStatus.Number,
		GatewayHistory: messageStatus.GatewayHistory,
	}
}

func (s *Session) portMessage(messageStatus *MessageStatus) {
	gateway, err := s.portGateway(messageStatus.GatewayHistory)
	if err != nil {
		metrics.PortedMessages.WithLabelValues("Failed", s.gateway, gateway).Inc()
		s.Write(&dtos.ReceiveLog{
			MessageID:  	 messageStatus.MessageID,
			SystemMessageID: messageStatus.SystemMessageID,
			Gateway:      	 s.gateway,
			MobileNo:     	 messageStatus.Number,
			MessageState: 	 "Porting Failed",
			ErrorCode:    	 "Unable to port message",
			Data:        	 messageStatus.Text,
		})

		logrus.Error(err)
		return
	}

	metrics.PortedMessages.WithLabelValues("Ported", s.gateway, gateway).Inc()

	s.portedStream.stream <- queue.MessageData{
		Id:             messageStatus.SystemMessageID,
		Gateway:        gateway,
		Sender:         messageStatus.Sender,
		Text:           messageStatus.Text,
		Number:         messageStatus.Number,
		GatewayHistory: messageStatus.GatewayHistory,
	}
}

func (s *Session) portGateway(history []string) (string, error) {
	alternatives := []string{"zain", "mobily"}

	switch len(history) {
	case 0:
		return "", errors.New("gateway history is empty, unable to port message")
	case 1:
		switch history[0] {
		case "stc":
			// If STC, return Zain
			return alternatives[0], nil
		case alternatives[0]:
			// If Zain, return Mobily
			return alternatives[1], nil
		case alternatives[1]:
			// If Mobily, return Zain
			return alternatives[0], nil
		}
	case 2:
		if contains(history, "stc") {
			return alternatives[1], nil
		}
		return "stc", nil
	case 3:
		metrics.TotallyFailedPortedMessages.WithLabelValues(history[0], history[1], history[2]).Inc()
		return "", errors.New("unable to port message, tried all gateways, len: 3")
	}
	return "", errors.New("invalid gateway history length: " + strconv.Itoa(len(history)))
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
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
