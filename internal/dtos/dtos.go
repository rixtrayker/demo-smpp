package dtos

import "github.com/phuslu/log"

type ReceiveLog struct{
	SystemMessageID     int64
	Gateway		 		string
	MessageID    		string   
	MessageState 		string   
	ErrorCode    		string   
	MobileNo     		string    
	Data         		string   
}

func (r *ReceiveLog) MarshalObject(e *log.Entry) {
	e.Str("Gateway", r.Gateway).Int64("SystemMessageID", r.SystemMessageID).Str("MessageID", r.MessageID).Str("MessageState", r.MessageState).Str("ErrorCode", r.ErrorCode).Str("MobileNo", r.MobileNo).Str("Data", r.Data)
}