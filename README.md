# Uptime-monitor
Uptime monitor is an application which monitors the uptime of systems.The application expects a url along with monitoring parameters. The inbound urls will be monitored continuosly by the application.

# Setup local environment 
* Install go on your machine https://golang.org/doc/install
* Install the necessary go packages by running the following commands
```
go get -u github.com/gin-gonic/gin
go get github.com/stretchr/testify/assert
go get github.com/gojektech/heimdall/httpclient
go get github.com/appleboy/gin-jwt/v2
go get github.com/DATA-DOG/go-sqlmock
go get github.com/go-test/deep
go get github.com/jinzhu/gorm
go get github.com/satori/go.uuid
```
* Setup mysql 
    * Install MySQL@5.7 on your machine https://dev.mysql.com/doc/refman/5.7/en/installing.html
    * Create a database named uptime and mention your username and password in Connect() function in database.go
