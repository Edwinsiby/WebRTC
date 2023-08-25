package main

import (
	"net/http"
	"rtc/pkg/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Static("/static", "./static")

	r.LoadHTMLGlob("templates/*")

	webRTCHandler := handlers.NewWebRTCHandler()

	webRTCHandler.SetupRoutes(r.Group("/v1"))

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.Run(":8080")
}
