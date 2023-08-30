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

	r.GET("/lobby", func(c *gin.Context) {
		c.HTML(http.StatusOK, "lobby.html", nil)
	})

	r.GET("/index", func(c *gin.Context) {
		room := c.DefaultQuery("room", "")
		c.HTML(http.StatusOK, "index.html", gin.H{"room": room})
	})

	r.GET("/get-offer", func(c *gin.Context) {
		c.String(http.StatusOK, handlers.GetStoredOffer())
	})

	r.Run(":8081")
}
