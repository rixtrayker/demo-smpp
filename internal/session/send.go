package session

import (
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/sirupsen/logrus"
)

func (s *Session) Send(sender, number, message string) error {
    submitSM := newSubmitSM(sender, number, message)
	ref := submitSM.SequenceNumber
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status[ref] = &MessageStatus{
		// DreamsMessageID: dreamsMessageId,
		Number: number,
	}

    s.outstandingCh <- struct{}{}
    err := s.transceiver.Transceiver().Submit(submitSM)
	logrus.WithField("message", message).Info("SubmitSM")
    if err != nil {
		logrus.WithError(err).Error("SubmitPDU error")
    }

    return nil
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

