package rdb

import (
	"app/lib/environment"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type ConnectionConfig struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	DB        string `json:"db"`
	Parameter string `json:"parameter"`
}

func getConnection(config ConnectionConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?%v", config.User, config.Password, config.Host, config.Port, config.DB, config.Parameter)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	db.Logger = db.Logger.LogMode(func() logger.LogLevel {
		if environment.IsDebug() {
			return logger.Info
		}

		return logger.Silent
	}())

	fmt.Printf("connection established. %v:%v %v \n", config.Host, config.Port, config.DB)
	return db, nil
}
