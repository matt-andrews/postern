package main

import (
	"bytes"
	"io"
	"log"
	"matt-andrews/postern/publishers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	publishers.Init()
	config := Discover("test.yaml")

	err := newRouter(config).Run("localhost:8080")
	if err != nil {
		panic(err)
	}
}

func run(c *gin.Context, match *ConfigMatch) {
	if err := publishers.PublishAll(match.Out, prepareBindings(c)); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
	})
}

func prepareBindings(c *gin.Context) map[string]any {
	result := make(map[string]any)

	//query params
	qry := make(map[string]string)
	for key, value := range c.Request.URL.Query() {
		qry[key] = value[0]
	}
	result["query"] = qry
	//

	//route params
	route := make(map[string]string)
	for _, param := range c.Params {
		route[param.Key] = param.Value
	}
	result["route"] = route
	//

	//headers
	headers := make(map[string]string)
	for name, values := range c.Request.Header {
		headers[name] = values[0] //probably want to string join instead
	}
	result["headers"] = headers
	//

	//request body as a raw string
	rawBody, err := c.GetRawData()
	if err != nil {
		panic("Unable to read body")
	}
	result["body"] = rawBody
	//reset the body so we can check json next
	c.Request.Body = io.NopCloser(bytes.NewReader(rawBody))
	//

	//request body as json if possible
	var payload any
	if err := c.ShouldBindJSON(&payload); err == nil {
		result["json"] = payload
	}
	//

	return result
}
