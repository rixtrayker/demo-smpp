package db

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/rixtrayker/demo-smpp/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var ErrDBNotConnected = errors.New("database connection not established")

var (
	DB     *gorm.DB
	once   sync.Once
)

func connect(ctx context.Context) error {
	var err error
	cfg := config.LoadConfig().DatabaseConfig

	if cfg == (config.DatabaseConfig{}) {
		return errors.New("DB config is empty")
	}

	once.Do(func() {
		dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&parseTime=true",
			cfg.User,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.DBName)

        
		DB, err = gorm.Open(mysql.Open(dataSourceName), &gorm.Config{})
		if err != nil {
			err = fmt.Errorf("failed to connect to database: %w", err)
			return
		}
		DB.WithContext(ctx)
		sqlDB, err := DB.DB()
		if err != nil {
			err = fmt.Errorf("failed to get underlying sql.DB: %w", err)
			return
		}
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetMaxOpenConns(70)
		fmt.Println("Successfully connected to database")
	})
	

	return err
}

func GetDBInstance(ctx context.Context) (*gorm.DB, error) {
	err := connect(ctx)
	if err != nil {
		return nil, err
	}
	return DB, nil
}

func Close(ctx context.Context) error {
	if DB == nil {
		return ErrDBNotConnected
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.Close()
}