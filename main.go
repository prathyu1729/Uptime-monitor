package main

import (
	"uptime/db"
	"uptime/handler"

	"github.com/gin-gonic/gin"
)

type urlinfo db.UrlInfo

func setupserver() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", pingEndpoint)
	return r
}

func pingEndpoint(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func main() {

	err := db.Connect()
	if err != nil {
		panic(err)
	}

	r := setupserver()
	m := make(map[int]handler.Channels)

	//api related functions
	r.POST("/urls/", handler.Posturl(m))
	r.GET("/urls/:id", handler.Geturlbyid())
	r.DELETE("/urls/:id", handler.Deleteurl(m))
	r.PATCH("/urls/:id", handler.Patchurl(m))
	r.POST("/urls/:id/activate", handler.Activateurl(m))
	r.POST("/urls/:id/deactivate", handler.Deactivateurl(m))

	//checking if data already exists in db
	urls := db.Getactiveurls()
	for _, url := range urls {
		id := int(url.ID)
		m[id] = handler.Channels{Quit: make(chan bool, 1), Data: make(chan db.Update, 1)}
		go handler.Monitor(url, m[id].Quit, m[id].Data)

	}
	//listening in the port 8080
	r.Run()

}
