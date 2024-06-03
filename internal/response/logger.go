package response

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Logger struct {
    logFilePattern string
    logFileMode    os.FileMode
    logFilePath    string
    logFile        *os.File
    logMutex       sync.Mutex
    logger         *logrus.Logger
    ctx            context.Context
    rotationFunc   func(time.Duration) error
}

var (
    instance     *Logger
    logMutex     sync.Mutex
    logFilePattern = "dlr_sms.2006-01-02.log"
    logFileMode    = os.FileMode(0644)
)

func NewLogger() (*Logger, error) {
    logMutex.Lock()
    defer logMutex.Unlock()
    if instance == nil {
        instance = &Logger{
            logFilePattern: logFilePattern,
            logFileMode:    logFileMode,
            ctx:           context.Background(),
        }
        err := instance.initLogger()
        if err != nil {
            return nil, err
        }
        go instance.autoRotateLogFile()
    }
    return instance, nil
}

func (l *Logger) initLogger() error {
    var err error
    l.logFilePath = filepath.Join(filepath.Dir(l.logFilePattern), time.Now().Format(l.logFilePattern))
    l.logFile, err = os.OpenFile(l.logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, l.logFileMode)
    if err != nil {
        fmt.Printf("Failed to open log file: %s\n", l.logFilePath)
        return err
    }
    l.logger = &logrus.Logger{
        Out:       l.logFile,
        Formatter: &logrus.JSONFormatter{},
        Level:     logrus.InfoLevel,
    }
    return nil
}

func (l *Logger) getLogger() *logrus.Logger {
    l.logMutex.Lock()
    defer l.logMutex.Unlock()
    if l.logger == nil {
        err := l.initLogger()
        if err != nil {
            fmt.Printf("Failed to initialize logger: %v\n", err)
            return nil
        }
    }
    return l.logger
}

func GetLogger() *logrus.Logger {
    logger, err := NewLogger()
    if err != nil {
        panic(err)
    }
    return logger.getLogger()
}

func (l *Logger) autoRotateLogFile() {
    if l.ctx == nil {
        return
    }
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()
    for {
        select {
        case <-l.ctx.Done():
            return
        case <-ticker.C:
            now := time.Now()
            nextDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
            duration := nextDay.Sub(now)
            if l.rotationFunc != nil {
                l.rotationFunc(duration)
            }
            l.logMutex.Lock()
            currentDate := time.Now().Format(l.logFilePattern)
            if currentDate != l.logFilePath {
                l.closeLogFile()
                l.logFilePath = filepath.Join(filepath.Dir(l.logFilePattern), currentDate)
                err := l.initLogger()
                if err != nil {
                    fmt.Printf("Failed to initialize logger during rotation: %v\n", err)
                }
            }
            l.logMutex.Unlock()
        }
    }
}

func (l *Logger) closeLogFile() {
    if l.logFile != nil {
        l.logFile.Close()
        l.logFile = nil
    }
}