package main

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	uuid "github.com/satori/go.uuid"
)

type Url struct {
	gorm.Model
	id                uuid.UUID `gorm:"type:uuid;primary_key;"`
	url               string
	crawl_timeout     time.Second
	frequency         uint
	failure_threshold uint
	status            string
	failure_count     uint
}

func Connect() {
	db, err := gorm.Open("mysql", "user:password@/dbname?charset=utf8&parseTime=True&loc=Local")
	defer db.Close()
}
