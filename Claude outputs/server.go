package main

import (
	"github.com/gin-gonic/gin"
)

// newRouter builds the HTTP handler for a config. main() serves it on a
// real port; tests serve it with httptest.NewServer.
func newRouter(config Config) *gin.Engine {
	router := gin.Default()

	for _, match := range config.Matches {
		router.Handle(match.Route.Method, match.Route.Path, func(c *gin.Context) {
			run(c, &match)
		})
	}

	return router
}
