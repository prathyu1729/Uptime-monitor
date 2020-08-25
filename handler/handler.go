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
	mu                = &sync.Mutex{}
	c                 = db.Caller{}
	dbGeturl          = c.Geturl
	dbInserturl       = c.Inserturl
	dbDeleteurl       = c.Deleteurl
	dbUpdatecrawl     = c.Updatecrawl
	dbUpdatefrequency = c.Updatefrequency
	dbUpdatethreshold = c.Updatethreshold
	dbActivateurl     = c.Activateurl
	dbDeactivateurl   = c.Deactivateurl
	dbUpdatefailure   = c.Updatefailure
)

// type sample struct{
// 	db db.Dbinteraction
// }

// func newsample(db db.Dbinteraction)*sample{
// 	return &sample{db:db}
// }

//function to create string to int, returns -1 for empty strings
func string_to_int(input string) int {
	if input == "" {
		return -1
	} else {
		result, _ := strconv.Atoi(input)
		return result
	}

}

//The go routine which will be created every time a new entry comes to the application
func Monitor(url db.UrlInfo, quit chan bool, data chan db.Update) {
	website := url.Url
	id := url.ID
	fmt.Printf("Monitor for %s created\n", website)
	crawl := url.Crawl_timeout
	timeout := time.Duration(crawl) * time.Second
	client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))
	frequency := url.Frequency
	ticker := time.NewTicker(time.Duration(frequency) * time.Second)
	threshold := url.Failure_threshold
	failure_count := 0
	for {
		select {
		case <-quit:
			fmt.Printf("Stoping monitor for %s\n", website)
			ticker.Stop()
			return
		case url_new := <-data:
			id = url_new.Id
			if url_new.Crawl_timeout != -1 {
				crawl = url_new.Crawl_timeout
				timeout = time.Duration(url_new.Crawl_timeout) * time.Second
				client = httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))
			}
			if url_new.Frequency != -1 {
				frequency = url_new.Frequency
				ticker = time.NewTicker(time.Duration(frequency) * time.Second)
			}
			if url_new.Failure_threshold != -1 {
				threshold = url_new.Failure_threshold
			}
			failure_count = 0
			url = db.UrlInfo{Url: website, Crawl_timeout: crawl, Frequency: frequency, Failure_count: 0, Status: "active", Failure_threshold: threshold}
			url.ID = id
		case t := <-ticker.C:
			fmt.Printf("%s Pinged at %s\n", website, t)
			res, err := client.Get(url.Url, nil)
			if err != nil || res.Status != "200 OK" {
				fmt.Printf("%s response failure\n", website)
				failure_count++
				dbUpdatefailure(url, failure_count)
				if failure_count >= threshold {
					_ = dbDeactivateurl(url)
					fmt.Printf("Stoping monitor for %s\n", website)
					return
				}

			} else {
				fmt.Printf("%s response success\n", website)
			}

		}
	}

}

//function to handle inserting new entries to the db
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
				"error": err,
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

//returns the url record, capturing id path variable from the url
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

//handles deletion of an existing url record in db
func Deleteurl(m map[string]Channels) func(*gin.Context) {
	return func(c *gin.Context) {

		id := c.Param("id")
		record, err := dbGeturl(id)
		dbDeleteurl(record)
		if err != nil {
			c.String(400, "")
		}
		mu.Lock()
		m[id].Quit <- true
		mu.Unlock()
		c.String(204, "success")
	}

}

//handles updation of fields of url records in db
func Patchurl(m map[string]Channels) func(*gin.Context) {
	return func(c *gin.Context) {

		id := c.Param("id")
		crawl_timeout := string_to_int(c.PostForm("crawl_timeout"))
		frequency := string_to_int(c.PostForm("frequency"))
		failure_threshold := string_to_int(c.PostForm("failure_threshold"))
		input := db.Update{Id: id, Crawl_timeout: crawl_timeout, Frequency: frequency, Failure_threshold: failure_threshold}
		var err1, err2, err3 error
		record, err := dbGeturl(id)
		if crawl_timeout != -1 {
			record, err1 = dbUpdatecrawl(record, crawl_timeout)
		}
		if frequency != -1 {
			record, err2 = dbUpdatefrequency(record, frequency)
		}
		if failure_threshold != -1 {
			record, err3 = dbUpdatethreshold(record, failure_threshold)
		}

		if err != nil || err1 != nil || err2 != nil || err3 != nil {
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

//handler to activate inactive url record
func Activateurl(m map[string]Channels) func(*gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		mu.Lock()
		rec, err := dbGeturl(id)
		err = dbActivateurl(rec)
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

//handler to deactivate an active url record
func Deactivateurl(m map[string]Channels) func(*gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")
		mu.Lock()
		rec, err := dbGeturl(id)
		err = dbDeactivateurl(rec)
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

//handler that returns all active urls
func Getactiveurls() ([]db.UrlInfo, error) {
	urls, err := c.Getactiveurls()
	return urls, err
}

//makes connection to the database
func Connecttodb() {
	err := c.Connect()
	_ = err
}

//closes the database
func Closedb() {

	c.Db.Close()
}
