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
# Execution
## Using a local copy
* Start the server by this command
```
go run main.go
```
## Using docker image
```
docker pull prathyu1729/uptime-monitor:init
docker run --rm -it -p 8080:8080 uptime-monitor
```
## Requests
* The server listens on port 8080.The following requests are expected by the application
```
POST /urls/
GET /urls/:id
DELETE /urls/:id
PATCH /urls/:id
POST /urls/:id/activate
POST /urls/:id/deactivate

```
* The following requests expect token authentication
```
POST /login
GET /auth/refresh_token
POST /auth/urls/
GET /auth/urls/:id
DELETE /auth/urls/:id
PATCH /auth/urls/:id
POST /auth/urls/:id/activate
POST /auth/urls/:id/deactivate
```
# Unit testing
* To run unit tests for http requests
```
cd handler/
go test
```
* To run unit tests for database interaction
```
cd db/
go test
```
