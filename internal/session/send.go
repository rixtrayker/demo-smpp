package session

import (
	"strconv"

	"github.com/linxGnu/gosmpp/data"
	"github.com/linxGnu/gosmpp/pdu"
	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/rixtrayker/demo-smpp/internal/queue"
	"github.com/sirupsen/logrus"
)

// ported redirect: zain -> mobily -> stc
var gateways = []string{"zain", "mobily", "stc"}

func (s *Session) Send(msg queue.MessageData) error {
    submitSM := newSubmitSM(msg.Sender, msg.Number, msg.Text)
    logrus.Info("Sending message to ", msg.Number)
    ref := submitSM.SequenceNumber

    gh := append(msg.GatewayHistory, s.gateway)

    s.mu.Lock()
    s.MessagesStatus[ref] = &MessageStatus{
        SystemMessageID: msg.Id,
        Sender:          msg.Sender,
        Text:            msg.Text,
        Number:          msg.Number,
        GatewayHistory:  gh,
    }

    s.mu.Unlock()

    if s.hasOutstanding {
        s.OutstandingCh <- struct{}{}
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

func (s *Session) handleSubmitSMResp(pd *pdu.SubmitSMResp) (pdu.PDU, bool) {
    select {
    case <-s.OutstandingCh:
        return s.processSubmitSMResp(pd)
    default:
        return s.processSubmitSMResp(pd)
    }
}

func (s *Session) processSubmitSMResp(pd *pdu.SubmitSMResp) (pdu.PDU, bool) {
    ref := pd.SequenceNumber
    s.mu.Lock()
    messageStatus := s.MessagesStatus[ref]
    s.mu.Unlock()

    errCode := strconv.Itoa(int(pd.CommandStatus))
    id := strconv.Itoa(messageStatus.SystemMessageID)

    s.responseWriter.WriteResponse(&dtos.ReceiveLog{
        MessageID:    id,
        Gateway:      s.gateway,
        MobileNo:     messageStatus.Number,
        MessageState: "Sent",
        ErrorCode:    errCode,
    })

    if(pd.CommandStatus == 0) {
        logrus.Info("SubmitSMResp Received")
    } else {
        if pd.CommandStatus == data.ESME_RINVDSTADR || s.gateway == "stc" {
            go s.portMessage(messageStatus)
        }
    }

    logrus.Info("SubmitSMResp Received")
    return pd.GetResponse(), false
}


func (s *Session) portMessage(messageStatus *MessageStatus) {    
    s.ResendChannel <- queue.MessageData{
        Id: messageStatus.SystemMessageID,
        Gateway: s.portGateway(messageStatus.GatewayHistory),
        Sender: messageStatus.Sender,
        Text: messageStatus.Text,
        Number: messageStatus.Number,
        GatewayHistory: messageStatus.GatewayHistory,
    }
}

// todo: history not including the result !!!!
func (s *Session) portGateway(history []string) string {
    // ["zain", "mobily", "stc"]
    if len(history) == 1 {
        return gateways[1]
    }

    if len(history) == 2 {
        return gateways[2]
    }

    return "Mobily"
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