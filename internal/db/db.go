package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sync"

	gormlogger "github.com/phuslu/log-contrib/gorm"

	"github.com/phuslu/log"
	"github.com/rixtrayker/demo-smpp/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var ErrDBNotConnected = errors.New("database connection not established")

var (
	DB     *gorm.DB
	once   sync.Once
	Logger logger.Interface
)

func connect() error {
	var err error
	cfg := config.LoadConfig("").DatabaseConfig

	if cfg == (config.DatabaseConfig{}) {
		return errors.New("DB config is empty")
	}

	once.Do(func() {
		Logger = gormlogger.New(&log.Logger{
			Level:      log.WarnLevel,
			TimeFormat: "15:04:05",
			Caller:     1,
			Writer: &log.FileWriter{
				Filename:   "db.log",
				MaxBackups: 14,
				LocalTime:  false,
			},
		}, logger.Config{
			SlowThreshold:             150 * time.Millisecond,
			LogLevel:                  logger.Warn,
		}, false)

		dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&parseTime=true",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DBName)

        
		DB, err = gorm.Open(mysql.Open(dataSourceName), &gorm.Config{
			Logger: Logger,
		})
		if err != nil {
			err = fmt.Errorf("failed to connect to database: %w", err)
			Logger.Error(context.Background(),err.Error())
			return
		}
		sqlDB, err := DB.DB()
		if err != nil {
			Logger.Error(context.Background(),fmt.Sprintf("failed to get underlying sql.DB: %v", err))
			return
		}
		sqlDB.SetMaxOpenConns(cfg.MaxConn)
		sqlDB.SetMaxIdleConns(cfg.MaxIdle)
		fmt.Println("Successfully connected to database")
	})
	

	return err
}

func GetDBInstance() (*gorm.DB, error) {
	err := connect()
	if err != nil {
		return nil, err
	}
	return DB, nil
}

func Close() error {
	if DB == nil {
		return ErrDBNotConnected
	}

	sqlDB, err := DB.DB()
	if err != nil {
		err = fmt.Errorf("failed to get underlying sql.DB: %w", err)
		Logger.Error(context.Background(),err.Error())
		return err
	}

	return sqlDB.Close()
}