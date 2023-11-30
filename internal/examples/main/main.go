package main

import (
	"github.com/polatbilek/gomigrator"
	_ "github.com/polatbilek/gomigrator/internal/examples/migrations"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

func getNow() time.Time {
	return time.Now().UTC()
}

func main() {
	dbConf := gorm.Config{
		SkipDefaultTransaction:   true,
		NowFunc:                  getNow,
		PrepareStmt:              false,
		DisableNestedTransaction: false,
		AllowGlobalUpdate:        true,
		QueryFields:              true,
		Dialector:                nil,
	}

	dsn := "YOUR_DSN_HERE"
	db, err := gorm.Open(mysql.Open(dsn), &dbConf)

	if err != nil {
		panic("Error while connecting database. Reason: " + err.Error())
	}

	gomigrator.Migrate("", db)
}
