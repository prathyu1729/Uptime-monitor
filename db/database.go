package db

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// type Base struct {
// 	ID        string `gorm:"type:uuid;primary_key;"`
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// 	DeletedAt *time.Time `sql:"index"`
// }

// func (base *Base) BeforeCreate(scope *gorm.Scope) error {
// 	uuid, err := uuid.NewV4()
// 	uuids := uuid.String()
// 	if err != nil {
// 		return err
// 	}
// 	return scope.SetColumn("ID", uuids)
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

type Caller struct {
	Db *gorm.DB
}

//	var info UrlInfo
//err := c.Db.Where("id = ?", id).Find(&info).Error
func (c *Caller) Deleteurl(id int) error {
	var info UrlInfo
	err := c.Db.Where("id = ?", id).Find(&info).Error
	c.Db.Delete(&info)
	return err
}

func (c *Caller) Activateurl(id int) error {
	var info UrlInfo
	err := c.Db.Where("id = ?", id).Find(&info).Error
	if err != nil {
		return err
	}
	if info.Status == "active" {
		return errors.New("url already active")
	}
	c.Db.Model(&info).Update("Status", "active")
	c.Db.Model(&info).Update("Failure_count", 0)
	return nil
}

func (c *Caller) Deactivateurl(id int) error {
	var info UrlInfo
	c.Db.Take(&info, id)
	if info.Status == "inactive" {
		return errors.New("url already inactive")
	}
	c.Db.Model(&info).Update("Status", "inactive")
	return nil

}

func (c *Caller) Updateurl(input Update) UrlInfo {
	var info UrlInfo
	id := input.Id
	c.Db.Take(&info, id)
	if input.Crawl_timeout != -1 {
		c.Db.Model(&info).Update("Crawl_timeout", input.Crawl_timeout)
	}
	if input.Frequency != -1 {
		c.Db.Model(&info).Update("Frequency", input.Frequency)
	}
	if input.Failure_threshold != -1 {
		c.Db.Model(&info).Update("Failure_threshold", input.Failure_threshold)
	}
	c.Db.Model(&info).Update("Failure_count", 0)

	return info

}

func (c *Caller) Updatefailure(id int, count int) {
	var info UrlInfo
	c.Db.Take(&info, id)
	c.Db.Model(&info).Update("Failure_count", count)

}

// func (c *Caller) Geturl(id int) (UrlInfo, error) {
// 	var info UrlInfo
// 	c.Db.First(&info, id)
// 	if info.ID == 0 {
// 		err := errors.New("record does not exist")
// 		return UrlInfo{}, err
// 	}
// 	return info, nil
// 	//err := c.Db.Where("id = ?", id).Find(info).Error
// }

func (c *Caller) Geturl(id int) (UrlInfo, error) {
	var info UrlInfo
	err := c.Db.Where("id = ?", id).Find(&info).Error
	return info, err
}

func (c *Caller) Getallurl() []UrlInfo {
	var urls []UrlInfo
	c.Db.Find(&urls)
	return urls
}

func (c *Caller) Getactiveurls() []UrlInfo {
	var urls []UrlInfo
	c.Db.Find(&urls, "status = ?", "active")
	return urls
}

func (c *Caller) Inserturl(record UrlInfo) UrlInfo {
	c.Db.Create(&record)
	var url UrlInfo
	c.Db.Last(&url)
	return url

}

func (c *Caller) Connect() error {
	var err error
	c.Db, err = gorm.Open("mysql", "prathyush:prathyush@/uptime?charset=utf8&parseTime=True&loc=Local")
	c.Db.AutoMigrate(&UrlInfo{})
	return err
}

func CreateInteractor(db *gorm.DB) Dbinteraction {
	return &Caller{
		Db: db,
	}
}
