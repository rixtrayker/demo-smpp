package session

import (
	"fmt"
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
        return s.send(submitSM)
    } else {
        return s.send(submitSM)
    }
}

func (s *Session) send(submitSM *pdu.SubmitSM) error{
    if s.smppSessions.transceiver != nil {
        return s.smppSessions.transceiver.Transceiver().Submit(submitSM)
    } else {
        return s.smppSessions.transmitter.Transmitter().Submit(submitSM)
    }
}

func (s *Session) handleSubmitSMResp(pd *pdu.SubmitSMResp) {
    select {
    case <-s.outstandingCh:
        s.processSubmitSMResp(pd)
    default:
        s.processSubmitSMResp(pd)
    }
    
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


func (s *Session) portMessage(messageStatus *MessageStatus) {    
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

// todo: history not including the result !!!!
func (s *Session) portGateway(history []string) (string, error){
    // ["zain", "mobily", "stc"]
    
    //default gateway
    if len(history) == 0 {
        return "", fmt.Errorf("gateway history, ported from void")
    }

    if len(history) == 1 && s.gateway != history[0] {
        return gateways[1], nil
    }

    if len(history) == 2 && s.gateway != history[0] && s.gateway != history[1] {
        return gateways[2], nil
    }

    return "", fmt.Errorf("unable to port message, tried all gateways, length: %d", len(history))
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