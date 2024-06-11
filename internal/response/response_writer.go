package response

import (
	"strconv"
	"sync"

	"github.com/rixtrayker/demo-smpp/internal/db"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/rixtrayker/demo-smpp/internal/models"
)


var responseWriter *Writer
var myDb *gorm.DB

type ResponseWriter interface {
	WriteResponse(msg *dtos.ReceiveLog)
	Close()
}

type Writer struct{
	Db *gorm.DB
	logger *logrus.Logger
	wg sync.WaitGroup
}

func NewResponseWriter() *Writer {
	if responseWriter == nil {
		logger := GetLogger()

		responseWriter = &Writer{
			// Db: getDB(),
			logger: logger,
		}
	}
	return responseWriter
}

func (w *Writer) WriteResponse(msg *dtos.ReceiveLog){
	w.wg.Add(1)
	go func(){
		defer w.wg.Done()
		w.writeDB(msg)
	}()
	
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.writeLog(msg)
	}()
}

func (w *Writer) writeLog(msg *dtos.ReceiveLog){
	loggingFields := logrus.Fields{
		"Gateway":      	msg.Gateway,
		"SystemMessageID:": msg.SystemMessageID,
		"MessageID":    	msg.MessageID,
		"MessageState": 	msg.MessageState,
		"ErrorCode":    	msg.ErrorCode,
		"MobileNo":     	msg.MobileNo,
		"Data":         	msg.Data,
	}

	w.logger.WithFields(loggingFields).Info("")
}

func (w *Writer) writeDB(msg *dtos.ReceiveLog){
	db := getDB()
	mobileNo, _  := strconv.ParseInt(msg.MobileNo, 10, 64)
	// todo: log error
	dlrSms := &models.DlrSms{
		SystemMessageID: msg.SystemMessageID,
		Gateway:      msg.Gateway,
		MessageID:    msg.MessageID,
		MessageState: msg.MessageState,
		ErrorCode:    msg.ErrorCode,
		MobileNo:     mobileNo,
		Data:         msg.Data,
	}
	
	tx := db.Create(&dlrSms)

	if tx.Error != nil {
		logrus.Printf("failed to insert data: %v\n", tx.Error)
	}
}

func (w* Writer) Close(){
	logrus.Info("closing response writer")
	w.wg.Wait()

	logrus.Info("closing logger")
	Close() // logger close
	logrus.Info("closing db")
	db.Close()
}


func getDB() *gorm.DB{
	var err error
	if myDb == nil {
		myDb, err = db.GetDBInstance()
	}

	if err != nil {
		logrus.WithError(err).Error("failed to get DB")
	}
	return myDb
}