package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gojektech/heimdall/httpclient"
	"github.com/jinzhu/gorm"
)

type Url struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {

	Connect()
	r := gin.Default()

	r.POST("/urls/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "post pong",
		})
	})

	fmt.Println("here")
	// listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	done := make(chan bool)

	//application pinging the urls and testing
	//read from the database
	//loop through all the urls

	var urls []string
	for _, url := range urls {
		_ = url
		go func() {

			timeout := 4 * time.Second
			client := httpclient.NewClient(httpclient.WithHTTPTimeout(timeout))
			ticker := time.NewTicker(6 * time.Second)
			failure_count := 0
			for {
				select {
				case <-done:
					ticker.Stop()
					return
				case t := <-ticker.C:
					fmt.Println("Tick at", t)
					res, err := client.Get("http://httpbin.org/delay/5", nil)
					if err != nil || res.Status != "200 OK" {
						fmt.Println("failure")
						failure_count++
					} else {
						fmt.Println("success")
					}
				}
			}

		}()

	}
	//end loop
	r.Run()
	//time.Sleep(20 * time.Second)
	//done <- true

}
