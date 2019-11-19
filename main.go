package main

import (
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mchmarny/gcputil/env"
	"github.com/mchmarny/gcputil/project"
)

var (
	//service
	logger      = log.New(os.Stdout, "", 0)
	projectID   = project.GetIDOrFail()
	port        = env.MustGetEnvVar("PORT", "8080")
	release     = env.MustGetEnvVar("RELEASE", "v0.0.1-default")
	accessToken = env.MustGetEnvVar("TOKEN", "")
	topicName   = env.MustGetEnvVar("TOPIC", "")
)

func main() {

	gin.SetMode(gin.ReleaseMode)

	// router
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// simple routes
	r.GET("/", defaultHandler)
	r.GET("/health", healthHandler)

	// api
	v1 := r.Group("/v1")
	{
		v1.POST("/notif", notifHandler)
	}

	// server
	hostPort := net.JoinHostPort("0.0.0.0", port)
	logger.Printf("Server starting: %s \n", hostPort)
	if err := r.Run(hostPort); err != nil {
		logger.Fatal(err)
	}
}
