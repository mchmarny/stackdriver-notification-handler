package main

import (
	"io/ioutil"
	"net/http"
	"strings"
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

func notifHandler(c *gin.Context) {
	var data []byte
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		logger.Printf("error capturing notification: %v", err)
	}
	logger.Println(string(data))

	// doing token after data for debugging
	token := strings.TrimSpace(c.Query("token"))
	if token != accessToken {
		logger.Printf("invalid access token. Got:%s Want:%s", token, accessToken)
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid access token",
			"status":  "Unauthorized",
		})
		return
	}

	if e := publish(c.Request.Context(), data); e != nil {
		logger.Printf("error publishing notification: %v", e)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error handling notification",
			"status":  "Failure",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification proccessed",
		"status":  "OK",
	})
}
