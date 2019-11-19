package main

import (
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/gin-gonic/gin"
)

func healthHandler(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

func defaultHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"release":      release,
		"request_on":   time.Now(),
		"request_from": c.Request.RemoteAddr,
	})
}

func printHandler(c *gin.Context) {
	rd, e := httputil.DumpRequest(c.Request, true)
	if e != nil {
		logger.Printf("error dumping request: %v", e)
	}
	logger.Println(string(rd))

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification proccessed",
		"status":  "OK",
	})
}
