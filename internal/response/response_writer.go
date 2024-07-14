package response

import (
	"strconv"
	"sync"

	"github.com/phuslu/log"
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
	logger log.Logger
	wg sync.WaitGroup
}

func NewResponseWriter() *Writer {
	if responseWriter == nil {
		responseWriter = &Writer{
			logger: log.Logger{
				Level: log.InfoLevel,
				Caller: 1,
				TimeFormat: "15:04:05",
				Writer: &log.FileWriter{
					Filename: "logs/dlr-sms/dlr-sms.log",
					MaxBackups: 30,
					LocalTime: false,
				},
			},
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
		// w.writeLog(msg)
		w.logger.Info().Object("DLR", msg).Msg("DLR Received")
	}()
}

// func (w *Writer) writeLog(msg *dtos.ReceiveLog){
// 	w.logger.Info().Object("DLR", msg).Msg("DLR Received")
// }

func (w *Writer) writeDB(msg *dtos.ReceiveLog){
	db := w.getDB()
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
		w.logger.Error().Err(tx.Error).Msg("failed to insert data")
		w.logger.Error().Object("DLR", msg).Msg("Error payload")
	}
}

func (w* Writer) Close(){
	logrus.Info("closing response writer")
	w.wg.Wait()

	logrus.Info("closing db")
	db.Close()
}


func (w *Writer) getDB() *gorm.DB{
	var err error
	if myDb == nil {
		myDb, err = db.GetDBInstance()
	}

	if err != nil {
		logrus.WithError(err).Error("failed to get DB")
		w.logger.Error().Err(err).Msg("failed to get DB")
		return nil
	}
	return myDb
}