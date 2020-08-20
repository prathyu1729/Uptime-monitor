package handler

import (
	"fmt"
	"strconv"
	"sync"
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
	mu              = &sync.Mutex{}
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
	website := url.Url
	id := url.ID
	fmt.Printf("Monitor for %s created\n", website)
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
			id = url_new.Id
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
			fmt.Printf("%s ", website)
			fmt.Println("monitoring at:", t)
			res, err := client.Get(url.Url, nil)
			if err != nil || res.Status != "200 OK" {
				fmt.Printf("%s failure\n", website)
				failure_count++
				dbUpdatefailure(id, failure_count)
				if failure_count >= threshold {
					_ = dbDeactivateurl(id)
					return
				}

			} else {
				fmt.Printf("%s success\n", website)
			}

		}
	}

}

func Posturl(m map[string]Channels) func(*gin.Context) {

	return func(c *gin.Context) {
		var err error
		url := c.PostForm("url")
		crawl_timeout, _ := strconv.Atoi(c.PostForm("crawl_timeout"))
		frequency, _ := strconv.Atoi(c.PostForm("frequency"))
		failure_threshold, _ := strconv.Atoi(c.PostForm("failure_threshold"))
		record := db.UrlInfo{Url: url, Crawl_timeout: crawl_timeout, Frequency: frequency, Failure_threshold: failure_threshold, Status: "active", Failure_count: 0}

		mu.Lock()
		record, err = dbInserturl(record)
		mu.Unlock()
		if err != nil {
			c.JSON(200, gin.H{
				"error": "url already exists",
			})
		} else {
			id := record.ID
			mu.Lock()
			m[id] = Channels{Quit: make(chan bool, 1), Data: make(chan db.Update, 1)}
			mu.Unlock()
			go Monitor(record, m[id].Quit, m[id].Data)
			c.JSON(200, gin.H{
				"Id":                id,
				"Url":               record.Url,
				"Crawl_timeout":     record.Crawl_timeout,
				"Frequency":         record.Frequency,
				"Failure_threshold": record.Failure_threshold,
				"Status":            record.Status,
				"Failure_count":     record.Failure_count,
			})
		}
	}
}

func Geturlbyid() func(*gin.Context) {
	return func(c *gin.Context) {

		id := c.Param("id")
		record, err := dbGeturl(id)
		if err != nil {
			c.JSON(400, gin.H{
				"message": "record does not exist",
			})
		} else {
			c.JSON(200, gin.H{
				"Id":                id,
				"Url":               record.Url,
				"Crawl_timeout":     record.Crawl_timeout,
				"Frequency":         record.Frequency,
				"Failure_threshold": record.Failure_threshold,
				"Status":            record.Status,
				"Failure_count":     record.Failure_count,
			})
		}
	}

}

func Deleteurl(m map[string]Channels) func(*gin.Context) {
	return func(c *gin.Context) {

		id := c.Param("id")
		mu.Lock()
		err := dbDeleteurl(id)
		mu.Unlock()
		if err != nil {
			c.String(400, "")
		}
		mu.Lock()
		m[id].Quit <- true
		mu.Unlock()
		c.String(204, "success")
	}

}

func Patchurl(m map[string]Channels) func(*gin.Context) {
	return func(c *gin.Context) {

		id := c.Param("id")
		crawl_timeout := string_to_int(c.PostForm("crawl_timeout"))
		frequency := string_to_int(c.PostForm("frequency"))
		failure_threshold := string_to_int(c.PostForm("failure_threshold"))
		input := db.Update{Id: id, Crawl_timeout: crawl_timeout, Frequency: frequency, Failure_threshold: failure_threshold}
		mu.Lock()
		record, err := dbUpdateurl(input)
		mu.Unlock()
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Could not update database",
			})
		}
		mu.Lock()
		m[id].Data <- input
		mu.Unlock()
		c.JSON(200, gin.H{
			"Id":                id,
			"Url":               record.Url,
			"Crawl_timeout":     record.Crawl_timeout,
			"Frequency":         record.Frequency,
			"Failure_threshold": record.Failure_threshold,
			"Status":            record.Status,
			"Failure_count":     record.Failure_count,
		})
	}

}

func Activateurl(m map[string]Channels) func(*gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		mu.Lock()
		err := dbActivateurl(id)
		mu.Unlock()
		if err != nil {
			c.JSON(400, gin.H{
				"error": "url already active",
			})
		} else {
			mu.Lock()
			record, _ := dbGeturl(id)
			mu.Unlock()
			record.Failure_count = 0
			go Monitor(record, m[id].Quit, m[id].Data)
			c.JSON(200, gin.H{
				"message": "update successful",
			})
		}
	}
}

func Deactivateurl(m map[string]Channels) func(*gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		mu.Lock()
		err := dbDeactivateurl(id)
		mu.Unlock()
		if err != nil {
			c.JSON(400, gin.H{
				"error": "url already inactive",
			})
		} else {
			mu.Lock()
			m[id].Quit <- true
			mu.Unlock()
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
