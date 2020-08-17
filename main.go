package main

import (
	"fmt"
	"strconv"
	"time"
	"uptime/db"

	"github.com/gin-gonic/gin"
	"github.com/gojektech/heimdall/httpclient"
)

type urlinfo db.UrlInfo

type channels struct {
	quit chan bool
	data chan db.Update
}

func string_to_int(input string) int {
	if input == "" {
		return -1
	} else {
		result, _ := strconv.Atoi(input)
		return result
	}

}

//var m sync.Mutex

func monitor(url db.UrlInfo, quit chan bool, data chan db.Update) {
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
			fmt.Println("monitoring at:", t)
			res, err := client.Get(url.Url, nil)
			if err != nil || res.Status != "200 OK" {
				fmt.Printf("%d failure\n", id)
				failure_count++
				db.Updatefailure(id, failure_count)
				if failure_count >= threshold {
					_ = db.Deactivateurl(id)
					return
				}

			} else {
				fmt.Printf("%d success\n", id)
			}

		}
	}

}

func main() {

	err := db.Connect()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("success")
	}
	r := gin.Default()
	m := make(map[int]channels)
	r.POST("/urls/", func(c *gin.Context) {

		url := c.PostForm("url")
		crawl_timeout, _ := strconv.Atoi(c.PostForm("crawl_timeout"))
		frequency, _ := strconv.Atoi(c.PostForm("frequency"))
		failure_threshold, _ := strconv.Atoi(c.PostForm("failure_threshold"))
		record := db.UrlInfo{Url: url, Crawl_timeout: crawl_timeout, Frequency: frequency, Failure_threshold: failure_threshold, Status: "active", Failure_count: 0}
		record = db.Inserturl(record)
		id := int(record.ID)
		m[id] = channels{quit: make(chan bool, 1), data: make(chan db.Update, 1)}
		go monitor(record, m[id].quit, m[id].data)
		c.JSON(200, gin.H{
			"url": url,
		})
	})

	r.GET("/urls/:id", func(c *gin.Context) {

		id, _ := strconv.Atoi(c.Param("id"))
		record, err := db.Geturl(id)
		if err != nil {
			c.JSON(500, gin.H{
				"message": "record does not exist",
			})
		} else {
			c.JSON(200, gin.H{
				"id":  id,
				"url": record.Url,
			})
		}
	})

	r.DELETE("/urls/:id", func(c *gin.Context) {

		id, _ := strconv.Atoi(c.Param("id"))
		err := db.Deleteurl(id)
		_ = err
		m[id].quit <- true
		c.String(204, "success")
	})

	r.PATCH("/urls/:id", func(c *gin.Context) {

		id, _ := strconv.Atoi(c.Param("id"))
		crawl_timeout := string_to_int(c.PostForm("crawl_timeout"))
		frequency := string_to_int(c.PostForm("frequency"))
		failure_threshold := string_to_int(c.PostForm("failure_threshold"))
		input := db.Update{Id: id, Crawl_timeout: crawl_timeout, Frequency: frequency, Failure_threshold: failure_threshold}
		db.Updateurl(input)
		m[id].data <- input
		c.JSON(200, gin.H{
			"crawl_time": crawl_timeout,
		})
	})

	r.POST("/urls/:id/activate", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		err := db.Activateurl(id)
		if err != nil {
			c.JSON(200, gin.H{
				"error": "url already active",
			})
		} else {
			record, _ := db.Geturl(id)
			record.Failure_count = 0
			go monitor(record, m[id].quit, m[id].data)
			c.JSON(200, gin.H{
				"message": "update successful",
			})
		}

	})

	r.POST("/urls/:id/deactivate", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		err := db.Deactivateurl(id)
		if err != nil {
			c.JSON(200, gin.H{
				"error": "url already inactive",
			})
		} else {
			m[id].quit <- true
			c.JSON(200, gin.H{
				"message": "update successful",
			})
		}

	})

	urls := db.Getactiveurls()
	for _, url := range urls {
		id := int(url.ID)
		fmt.Printf("here:%d\n", id)
		m[id] = channels{quit: make(chan bool, 1), data: make(chan db.Update, 1)}
		go monitor(url, m[id].quit, m[id].data)

	}

	r.Run()

}
