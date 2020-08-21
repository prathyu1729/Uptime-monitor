package main

import (
	"log"
	"time"
	"uptime/db"
	"uptime/handler"

	jwt "github.com/appleboy/gin-jwt/v2"
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

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var identityKey = "id"

func helloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	user, _ := c.Get(identityKey)
	c.JSON(200, gin.H{
		"userID":   claims[identityKey],
		"userName": user.(*User).UserName,
		"text":     "Hello World.",
	})
}

// User demo
type User struct {
	UserName  string
	FirstName string
	LastName  string
}

func main() {

	handler.Connecttodb()

	r := setupserver()
	defer handler.Closedb()

	//map to store the channels for each goroutine to be created with its id as the key
	m := make(map[string]handler.Channels)

	//api related functions
	r.POST("/urls/", handler.Posturl(m))
	r.GET("/urls/:id", handler.Geturlbyid())
	r.DELETE("/urls/:id", handler.Deleteurl(m))
	r.PATCH("/urls/:id", handler.Patchurl(m))
	r.POST("/urls/:id/activate", handler.Activateurl(m))
	r.POST("/urls/:id/deactivate", handler.Deactivateurl(m))

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// the jwt middleware for token authentication
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*User); ok {
				return jwt.MapClaims{
					identityKey: v.UserName,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &User{
				UserName: claims[identityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			userID := loginVals.Username
			password := loginVals.Password

			if (userID == "admin" && password == "admin") || (userID == "test" && password == "test") {
				return &User{
					UserName:  userID,
					LastName:  "Bo-Yi",
					FirstName: "Wu",
				}, nil
			}

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*User); ok && v.UserName == "admin" {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	r.POST("/login", authMiddleware.LoginHandler)

	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	auth := r.Group("/auth")
	// Refresh time can be longer than token timeout
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/hello", helloHandler)
		auth.POST("/urls/", handler.Posturl(m))
		auth.GET("/urls/:id", handler.Geturlbyid())
		auth.DELETE("/urls/:id", handler.Deleteurl(m))
		auth.PATCH("/urls/:id", handler.Patchurl(m))
		auth.POST("/urls/:id/activate", handler.Activateurl(m))
		auth.POST("/urls/:id/deactivate", handler.Deactivateurl(m))
	}

	//checking if data already exists in db
	urls := handler.Getactiveurls()
	for _, url := range urls {
		id := url.ID
		m[id] = handler.Channels{Quit: make(chan bool, 1), Data: make(chan db.Update, 1)}
		go handler.Monitor(url, m[id].Quit, m[id].Data)

	}
	//listening in the port 8080
	r.Run()

}
