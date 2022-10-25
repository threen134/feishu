package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("start signing server...")
	router := gin.Default()

	// forwarding webhook
	router.POST("/open-apis/bot/v2/hook/1ef26cc4-a7e6-4295-8483-3ac8e923356e", forwarding)

	router.Run()
}