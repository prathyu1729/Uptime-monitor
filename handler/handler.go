package handler

import (
	"fmt"
	"strconv"
	"time"
	"uptime/db"

	"github.com/gin-gonic/gin"
	"github.com/gojektech/heimdall/httpclient"
)

type Channels struct {
	Quit chan bool
	Data chan db.Update
}

var (
	c               = db.Caller{}
	dbGeturl        = c.Geturl
	dbInserturl     = c.Inserturl
	dbDeleteurl     = c.Deleteurl
	dbUpdateurl     = c.Updateurl
	dbActivateurl   = c.Activateurl
	dbDeactivateurl = c.Deactivateurl
	dbUpdatefailure = c.Updatefailure
)

func string_to_int(input string) int {
	if input == "" {
		return -1
	} else {
		result, _ := strconv.Atoi(input)
		return result
	}

}

func Monitor(url db.UrlInfo, quit chan bool, data chan db.Update) {
	id := int(url.ID)
	fmt.Printf("%d created\n", id)
	timeout := time.Duration(url.Crawl_timeout) * time.Second
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))
	ticker := time.NewTicker(time.Duration(url.Frequency) * time.Second)
	threshold := url.Failure_threshold
	failure_count := 0
	for {
		select {
		case <-quit:
			fmt.Printf("deleting")
			ticker.Stop()
			return
		case url_new := <-data:
			fmt.Printf("data received %d", url_new.Failure_threshold)
			id = int(url_new.Id)
			if url_new.Crawl_timeout != -1 {
				timeout = time.Duration(url_new.Crawl_timeout) * time.Second
				client = httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))
			}
			if url_new.Frequency != -1 {
				ticker = time.NewTicker(time.Duration(url_new.Frequency) * time.Second)
			}
			if url_new.Failure_threshold != -1 {
				threshold = url_new.Failure_threshold
			}
			failure_count = 0

		case t := <-ticker.C:
			fmt.Printf("%d ", id)
			fmt.Println("monitoring at:", t)
			res, err := client.Get(url.Url, nil)
			if err != nil || res.Status != "200 OK" {
				fmt.Printf("%d failure\n", id)
				failure_count++
				dbUpdatefailure(id, failure_count)
				if failure_count >= threshold {
					_ = dbDeactivateurl(id)
					return
				}

			} else {
				fmt.Printf("%d success\n", id)
			}

		}
	}

}

func Posturl(m map[int]Channels) func(*gin.Context) {
	return func(c *gin.Context) {

		url := c.PostForm("url")
		crawl_timeout, _ := strconv.Atoi(c.PostForm("crawl_timeout"))
		frequency, _ := strconv.Atoi(c.PostForm("frequency"))
		failure_threshold, _ := strconv.Atoi(c.PostForm("failure_threshold"))
		record := db.UrlInfo{Url: url, Crawl_timeout: crawl_timeout, Frequency: frequency, Failure_threshold: failure_threshold, Status: "active", Failure_count: 0}
		record = dbInserturl(record)
		id := int(record.ID)
		m[id] = Channels{Quit: make(chan bool, 1), Data: make(chan db.Update, 1)}
		go Monitor(record, m[id].Quit, m[id].Data)
		c.JSON(200, gin.H{
			"url": url,
		})
	}
}

func Geturlbyid() func(*gin.Context) {
	return func(c *gin.Context) {

		id, _ := strconv.Atoi(c.Param("id"))
		record, err := dbGeturl(id)
		_ = record
		if err != nil {
			c.JSON(500, gin.H{
				"message": "record does not exist",
			})
		} else {
			c.JSON(200, gin.H{
				"id": id,
				//"url": record.Url,
			})
		}
	}

}

func Deleteurl(m map[int]Channels) func(*gin.Context) {
	return func(c *gin.Context) {

		id, _ := strconv.Atoi(c.Param("id"))
		err := dbDeleteurl(id)
		_ = err
		m[id].Quit <- true
		c.String(204, "success")
	}

}

func Patchurl(m map[int]Channels) func(*gin.Context) {
	return func(c *gin.Context) {

		id, _ := strconv.Atoi(c.Param("id"))
		crawl_timeout := string_to_int(c.PostForm("crawl_timeout"))
		frequency := string_to_int(c.PostForm("frequency"))
		failure_threshold := string_to_int(c.PostForm("failure_threshold"))
		input := db.Update{Id: id, Crawl_timeout: crawl_timeout, Frequency: frequency, Failure_threshold: failure_threshold}
		record := dbUpdateurl(input)
		m[id].Data <- input
		c.JSON(200, gin.H{
			"ID":                id,
			"Url":               record.Url,
			"Crawl_timeout":     crawl_timeout,
			"Frequency":         frequency,
			"Failure_threshold": failure_threshold,
			"Status":            record.Status,
			"Failure_count":     record.Failure_count,
		})
	}

}

func Activateurl(m map[int]Channels) func(*gin.Context) {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		err := dbActivateurl(id)
		if err != nil {
			c.JSON(200, gin.H{
				"error": "url already active",
			})
		} else {
			record, _ := dbGeturl(id)
			record.Failure_count = 0
			go Monitor(record, m[id].Quit, m[id].Data)
			c.JSON(200, gin.H{
				"message": "update successful",
			})
		}
	}
}

func Deactivateurl(m map[int]Channels) func(*gin.Context) {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		err := dbDeactivateurl(id)
		if err != nil {
			c.JSON(200, gin.H{
				"error": "url already inactive",
			})
		} else {
			m[id].Quit <- true
			c.JSON(200, gin.H{
				"message": "update successful",
			})
		}

	}

}

func Getactiveurls() []db.UrlInfo {
	urls := c.Getactiveurls()
	return urls
}

func Connecttodb() {
	err := c.Connect()
	_ = err
}

func Closedb() {

	c.Db.Close()
}
