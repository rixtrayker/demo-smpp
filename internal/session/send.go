package session

import (
	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/sirupsen/logrus"
)

func (s *Session) Send(sender, number, message string) error {
    submitSM := newSubmitSM(sender, number, message)
    logrus.Info("Sending message to ", number)
    ref := submitSM.SequenceNumber
    s.mu.Lock()
    s.Status[ref] = &MessageStatus{
        Number: number,
    }
    s.mu.Unlock()

    if s.hasOutstanding {
        s.outstandingCh <- struct{}{}
        return s.send(submitSM)
    } else {
        return s.send(submitSM)
    }
}

func (s *Session) send(submitSM *pdu.SubmitSM) error{
    if s.transceiver != nil {
        return s.transceiver.Transceiver().Submit(submitSM)
    } else {
        return s.transmitter.Transmitter().Submit(submitSM)
    }
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