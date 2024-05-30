package response

import (
	"context"

	"github.com/rixtrayker/demo-smpp/internal/db"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/rixtrayker/demo-smpp/internal/models"
)


var responseWriter *Writer
var myDb *gorm.DB

type ResponseWriter interface {
	WriteResponse(msg dtos.ReceiveLog)
}

type Writer struct{
	ctx context.Context
	Db *gorm.DB
}

func NewResponseWriter(ctx context.Context) *Writer {

	if responseWriter == nil {
		responseWriter = &Writer{
			ctx: ctx,
			Db: getDB(ctx),
		}
	}
	return responseWriter
}

func (w *Writer) WriteResponse(msg dtos.ReceiveLog){
	go func(){
		w.writeDB(msg)
	}()

	go func() {
		w.writeLog(msg)
	}()
}

func (w *Writer) writeLog(msg dtos.ReceiveLog){
	loggingFields := logrus.Fields{
		"MessageID":    msg.MessageID,
		"MessageState": msg.MessageState,
		"ErrorCode":    msg.ErrorCode,
		"MobileNo":     msg.MobileNo,
		"CurrentTime":  msg.CurrentTime,
		"Data":         msg.Data,
	}

	logrus.WithFields(loggingFields).Info(" new dlr ")
}

func (w *Writer) writeDB(msg dtos.ReceiveLog){
	db := getDB(w.ctx)
	dlrSms := &models.DlrSms{
		MessageID:    msg.MessageID,
		MessageState: msg.MessageState,
		ErrorCode:    msg.ErrorCode,
		MobileNo:     msg.MobileNo,
		CurrentTime:  msg.CurrentTime,
		Data:         msg.Data,
	}
	
	tx := db.Create(&dlrSms)

	if tx.Error != nil {
		logrus.Printf("failed to insert data: %v\n", tx.Error)
	}
}

// should I use context ?

func getDB(ctx context.Context) *gorm.DB{
	var err error
	if myDb == nil {
		myDb, err = db.GetDBInstance(ctx)
	}

	if err != nil {
		logrus.WithError(err).Error("failed to get DB")
	}
	return myDb
}