package db

import (
	"github.com/rixtrayker/demo-smpp/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Init(cfg config.Config) *gorm.DB {
    db, err := gorm.Open(mysql.Open(cfg.DatabaseConfig.DSN), &gorm.Config{})
    if err != nil {
        panic("failed to connect to database")
    }
    return db
}
