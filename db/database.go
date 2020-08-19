package db

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// type Model struct {
// 	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// 	DeletedAt *time.Time `sql:"index"`
// }
var db *gorm.DB

type UrlInfo struct {
	gorm.Model
	Url               string
	Crawl_timeout     int
	Frequency         int
	Failure_threshold int
	Status            string
	Failure_count     int
}

type Update struct {
	Id                int
	Crawl_timeout     int
	Frequency         int
	Failure_threshold int
}

type Dbinteraction interface {
	Deleteurl(id int) error
	Activateurl(id int) error
	Deactivateurl(id int) error
	Updateurl(input Update) UrlInfo
	Updatefailure(id int, count int)
	Geturl(id int) (UrlInfo, error)
	Getallurl() []UrlInfo
	Getactiveurls() []UrlInfo
	Inserturl(record UrlInfo) UrlInfo
	Connect() error
}

func Deleteurl(id int) error {
	db, err := gorm.Open("mysql", "prathyush:prathyush@/uptime?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("Could not connect to database")
	}
	defer db.Close()
	var info UrlInfo
	db.Take(&info, id)
	// if info.Url == nil {
	// 	return errors.New("record does not exist")
	// }
	db.Delete(&info)
	return nil
}

func Activateurl(id int) error {
	db, err := gorm.Open("mysql", "prathyush:prathyush@/uptime?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("Could not connect to database")
	}
	defer db.Close()
	var info UrlInfo
	db.Take(&info, id)
	if info.Status == "active" {
		return errors.New("url already active")
	}
	db.Model(&info).Update("Status", "active")
	db.Model(&info).Update("Failure_count", 0)
	return nil
}

func Deactivateurl(id int) error {
	db, err := gorm.Open("mysql", "prathyush:prathyush@/uptime?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("Could not connect to database")
	}
	defer db.Close()
	var info UrlInfo
	db.Take(&info, id)
	if info.Status == "inactive" {
		return errors.New("url already inactive")
	}
	db.Model(&info).Update("Status", "inactive")
	return nil

}

func Updateurl(input Update) UrlInfo {
	db, err := gorm.Open("mysql", "prathyush:prathyush@/uptime?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("Could not connect to database")
	}
	defer db.Close()
	var info UrlInfo
	id := input.Id
	db.Take(&info, id)
	if input.Crawl_timeout != -1 {
		db.Model(&info).Update("Crawl_timeout", input.Crawl_timeout)
	}
	if input.Frequency != -1 {
		db.Model(&info).Update("Frequency", input.Frequency)
	}
	if input.Failure_threshold != -1 {
		db.Model(&info).Update("Failure_threshold", input.Failure_threshold)
	}
	db.Model(&info).Update("Failure_count", 0)

	return info

}

func Updatefailure(id int, count int) {
	db, err := gorm.Open("mysql", "prathyush:prathyush@/uptime?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("Could not connect to database")
	}
	defer db.Close()
	var info UrlInfo
	db.Take(&info, id)
	db.Model(&info).Update("Failure_count", count)

}

func Geturl(id int) (UrlInfo, error) {
	db, err := gorm.Open("mysql", "prathyush:prathyush@/uptime?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("Could not connect to database")
	}
	defer db.Close()
	var info UrlInfo
	db.Take(&info, id)
	if info.ID == 0 {
		err := errors.New("record does not exist")
		return UrlInfo{}, err
	}
	return info, nil
}

func Getallurl() []UrlInfo {
	db, err := gorm.Open("mysql", "prathyush:prathyush@/uptime?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("Could not connect to database")
	}
	defer db.Close()
	var urls []UrlInfo
	db.Find(&urls)
	return urls
}

func Getactiveurls() []UrlInfo {
	db, err := gorm.Open("mysql", "prathyush:prathyush@/uptime?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("Could not connect to database")
	}
	defer db.Close()
	var urls []UrlInfo
	db.Find(&urls, "status = ?", "active")
	return urls
}

func Inserturl(record UrlInfo) UrlInfo {
	db, err := gorm.Open("mysql", "prathyush:prathyush@/uptime?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("Could not connect to database")
	}
	defer db.Close()
	db.Create(&record)
	var url UrlInfo
	db.Last(&url)
	return url

}

func Connect() error {
	var err error
	db, err = gorm.Open("mysql", "prathyush:prathyush@/uptime?charset=utf8&parseTime=True&loc=Local")
	defer db.Close()
	db.AutoMigrate(&UrlInfo{})
	return err
}
