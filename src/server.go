package main

import (
	"github.com/gin-gonic/gin"
)

// builds the HTTP handler for a config
func newRouter(config Config) *gin.Engine {
	router := gin.Default()

	for _, match := range config.Matches {
		router.Handle(match.Route.Method, match.Route.Path, func(c *gin.Context) {
			run(c, &match)
		})
	}

	return router
}
