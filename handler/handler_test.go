package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uptime/db"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func server() *gin.Engine {
	r := gin.Default()
	m := make(map[string]Channels)
	m["2"] = Channels{Quit: make(chan bool, 1), Data: make(chan db.Update, 1)}
	r.POST("/urls/", Posturl(m))
	r.PATCH("/urls/:id", Patchurl(m))
	r.GET("/urls/:id", Geturlbyid())
	r.DELETE("/urls/:id", Deleteurl(m))
	r.POST("/urls/:id/activate", Activateurl(m))
	r.POST("/urls/:id/deactivate", Deactivateurl(m))
	return r
}

func performRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestDeactivateurl(t *testing.T) {
	dbDeactivateurl = func(info db.UrlInfo) error {
		return nil
	}
	router := server()
	w := performRequest(router, "POST", "/urls/2/deactivate", nil)
	assert.Equal(t, http.StatusOK, w.Code)

}

func TestActivateurl(t *testing.T) {
	dbActivateurl = func(info db.UrlInfo) error {
		return nil

	}

	dbGeturl = func(i string) (db.UrlInfo, error) {
		record := db.UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 15, Failure_threshold: 2, Status: "active", Failure_count: 0}
		record.ID = i
		return record, nil
	}

	router := server()
	w := performRequest(router, "POST", "/urls/2/activate", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteurl(t *testing.T) {

	dbDeleteurl = func(info db.UrlInfo) error {
		//_ = id
		return nil
	}
	router := server()
	w := performRequest(router, "DELETE", "/urls/2", nil)
	assert.Equal(t, 204, w.Code)

}

func TestPosturl(t *testing.T) {

	//newsample()
	dbInserturl = func(r db.UrlInfo) (db.UrlInfo, error) {
		record := db.UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 15, Failure_threshold: 2, Status: "active", Failure_count: 0}
		record.ID = "1"
		return record, nil

	}
	body := gin.H{
		"url": "abc.com",
	}
	content := strings.NewReader("url=abc.com&crawl_timeout=10&frequency=15&failure_threshold=2;empty=&")
	router := server()
	w := performRequest(router, "POST", "/urls/", content)
	assert.Equal(t, http.StatusOK, w.Code)
	var response db.UrlInfo
	err := json.Unmarshal([]byte(w.Body.String()), &response)
	value := response.Url
	assert.Nil(t, err)
	//assert.True(t, exists)
	assert.Equal(t, body["url"], value)

}

func TestGeturlbyid(t *testing.T) {

	dbGeturl = func(i string) (db.UrlInfo, error) {
		record := db.UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 15, Failure_threshold: 2, Status: "active", Failure_count: 0}
		record.ID = i
		return record, nil
	}
	router := server()
	w := performRequest(router, "GET", "/urls/1", nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var response db.UrlInfo
	err := json.Unmarshal([]byte(w.Body.String()), &response)
	value := response.ID
	assert.Nil(t, err)
	assert.Equal(t, "1", value)

}

func TestPatchurl(t *testing.T) {

	dbGeturl = func(i string) (db.UrlInfo, error) {
		record := db.UrlInfo{Url: "abc.com", Crawl_timeout: 10, Frequency: 15, Failure_threshold: 2, Status: "active", Failure_count: 0}
		record.ID = i
		return record, nil
	}

	dbUpdatecrawl = func(record db.UrlInfo, crawl int) (db.UrlInfo, error) {
		record.Crawl_timeout = crawl
		return record, nil
	}

	dbUpdatefrequency = func(record db.UrlInfo, f int) (db.UrlInfo, error) {
		record.Frequency = f
		return record, nil
	}

	dbUpdatefrequency = func(record db.UrlInfo, f int) (db.UrlInfo, error) {
		record.Frequency = f
		return record, nil
	}

	dbUpdatethreshold = func(record db.UrlInfo, t int) (db.UrlInfo, error) {
		record.Failure_threshold = t
		return record, nil
	}

	router := server()
	content := strings.NewReader("crawl_timeout=11&frequency=16&failure_threshold=3;empty=&")
	w := performRequest(router, "PATCH", "/urls/2", content)
	assert.Equal(t, http.StatusOK, w.Code)
	var response db.UrlInfo
	err := json.Unmarshal([]byte(w.Body.String()), &response)
	id := response.ID
	url := response.Url
	crawl_timeout := response.Crawl_timeout
	frequency := response.Frequency
	status := response.Status
	failure_count := response.Failure_count
	failure_threshold := response.Failure_threshold
	assert.Nil(t, err)
	assert.Equal(t, "2", id)
	assert.Equal(t, "abc.com", url)
	assert.Equal(t, 11, crawl_timeout)
	assert.Equal(t, 16, frequency)
	assert.Equal(t, "active", status)
	assert.Equal(t, 0, failure_count)
	assert.Equal(t, 3, failure_threshold)

}
