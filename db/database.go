package db

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	uuid "github.com/satori/go.uuid"
)

type Base struct {
	ID        string `gorm:"type:char(36);primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time //`gorm:"default:null"`
}

func (base *Base) BeforeCreate(scope *gorm.Scope) error {
	uuid, err := uuid.NewV4()
	uuids := uuid.String()
	if err != nil {
		return err
	}
	return scope.SetColumn("ID", uuids)
}

var db *gorm.DB

type UrlInfo struct {
	Base
	Url               string `gorm:"unique;not null"`
	Crawl_timeout     int    `gorm:"not null"`
	Frequency         int    `gorm:"not null"`
	Failure_threshold int    `gorm:"not null"`
	Status            string `gorm:"not null"`
	Failure_count     int    //`gorm:"unique;not null"
}

type Update struct {
	Id                string
	Crawl_timeout     int
	Frequency         int
	Failure_threshold int
}

type Dbinteraction interface {
	Deleteurl(info UrlInfo) error
	Activateurl(info UrlInfo) error
	Deactivateurl(info UrlInfo) error
	//Updateurl(input Update) (UrlInfo, error)
	Updatefailure(info UrlInfo, count int) (UrlInfo, error)
	Geturl(id string) (UrlInfo, error)
	Getallurl() []UrlInfo
	Getactiveurls() ([]UrlInfo, error)
	Inserturl(record UrlInfo) (UrlInfo, error)
	Updatecrawl(c UrlInfo, crawl int) (UrlInfo, error)
	Updatefrequency(c UrlInfo, crawl int) (UrlInfo, error)
	Updatethreshold(c UrlInfo, crawl int) (UrlInfo, error)
	Connect() error
}

type Caller struct {
	Db *gorm.DB
}

//tested
func (c *Caller) Geturl(id string) (UrlInfo, error) {
	var info UrlInfo
	err := c.Db.Where("id = ?", id).Find(&info).Error
	return info, err
}

//tested
func (c *Caller) Deleteurl(info UrlInfo) error {
	r := c.Db.Delete(&info)
	return r.Error

}

//tested
func (c *Caller) Activateurl(info UrlInfo) error {

	if info.Status == "active" {
		return errors.New("url already active")
	}
	c.Db.Model(&info).Update("Status", "active")
	return nil
}

//tested
func (c *Caller) Deactivateurl(info UrlInfo) error {

	if info.Status == "inactive" {
		return errors.New("url already inactive")
	}
	c.Db.Model(&info).Update("Status", "inactive")
	return nil

}

//tested
func (c *Caller) Updatecrawl(record UrlInfo, crawl int) (UrlInfo, error) {
	err := c.Db.Model(&record).Update("Crawl_timeout", crawl)
	return record, err.Error

}

//tested
func (c *Caller) Updatefrequency(record UrlInfo, f int) (UrlInfo, error) {
	err := c.Db.Model(&record).Update("Frequency", f)
	return record, err.Error

}

//tested
func (c *Caller) Updatethreshold(record UrlInfo, t int) (UrlInfo, error) {
	err := c.Db.Model(&record).Update("Failure_threshold", t)
	return record, err.Error

}

//tested
func (c *Caller) Updatefailure(info UrlInfo, count int) (UrlInfo, error) {
	err := c.Db.Model(&info).Update("Failure_count", count)
	return info, err.Error

}

func (c *Caller) Getallurl() []UrlInfo {
	var urls []UrlInfo
	c.Db.Find(&urls)
	return urls
}

//tested
func (c *Caller) Getactiveurls() ([]UrlInfo, error) {
	var urls []UrlInfo
	er := c.Db.Find(&urls, "status = ?", "active")
	return urls, er.Error
}

//tested
func (c *Caller) Inserturl(record UrlInfo) (UrlInfo, error) {
	er := c.Db.Create(&record)
	return record, er.Error

}

func (c *Caller) Connect() error {
	var err error
	c.Db, err = gorm.Open("mysql", "prathyush:prathyush@(localhost:3306)/uptime?charset=utf8&parseTime=True&loc=Local")
	c.Db.AutoMigrate(&UrlInfo{})
	return err
}

func CreateInteractor(db *gorm.DB) Dbinteraction {
	return &Caller{
		Db: db,
	}
}
