package runtime

import (
	"bytes"
	"flag"
	"io"
	"log"
	"matt-andrews/postern/publishers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Start() {
	publishers.Init()
	config := getArgSuppliedDir()

	err := NewRouter(config).Run(getRunAddr())
	if err != nil {
		panic(err)
	}
}

// builds the HTTP handler for a config
func NewRouter(config Config) *gin.Engine {
	router := gin.Default()

	for _, match := range config.Matches {
		router.Handle(match.Route.Method, match.Route.Path, func(c *gin.Context) {
			run(c, &match)
		})
	}

	return router
}

func getArgSuppliedDir() Config {
	dir := flag.String("dir", "../scripts/test.yaml", "directory or single config file to load")
	flag.Parse()
	return Discover(*dir)
}

func getRunAddr() string {
	return "localhost:8080"
}

func run(c *gin.Context, match *ConfigMatch) {
	if err := publishers.PublishAll(match.Out, prepareBindings(c)); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.Status(200)
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
	result["body"] = string(rawBody)
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
