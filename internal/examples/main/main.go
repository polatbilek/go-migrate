package main

import (
	_ "go-migrate/internal/examples/migrations"
	"go-migrate/migrator"
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

	dsn := "api_user:9f81c7772b9d1e6b0b1074a0cff8b431@tcp(localhost:3306)/facebookapp?charset=utf8mb4&parseTime=True&loc=UTC"
	db, err := gorm.Open(mysql.Open(dsn), &dbConf)

	if err != nil {
		panic("Error while connecting database. Reason: " + err.Error())
	}

	migrator.Migrate("", db)
	time.Sleep(1000000)
}
