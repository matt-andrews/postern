package main

import (
	"matt-andrews/harbinger/publishers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	publishers.Init()
	config := Discover("test.yaml")
	//fmt.Printf("%+v\n", config)
	router := gin.Default()

	for _, match := range config.Matches {
		router.Handle(match.Route.Method, match.Route.Path, func(c *gin.Context) {
			run(c, &match)
		})
	}

	router.Run("localhost:8080")
}

func run(c *gin.Context, match *ConfigMatch) {
	publishers.PublishAll(match.Out)
	c.JSON(http.StatusOK, gin.H{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
	})
}
