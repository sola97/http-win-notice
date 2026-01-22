package model

import (
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"http-win-notice/utils/setting"
)

var db *gorm.DB
var err error

func InitDb() {
	db, err = gorm.Open(sqlite.Open(setting.DbPath), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Msg{})
}

func PingDb() error {
	if db == nil {
		return errors.New("db not initialized")
	}
	return db.Exec("SELECT 1").Error
}
