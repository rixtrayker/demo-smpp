package dtos

type ReceiveLog struct{
	SystemMessageID     int64
	Gateway		 		string
	MessageID    		string   
	MessageState 		string   
	ErrorCode    		string   
	MobileNo     		string    
	Data         		string   
}