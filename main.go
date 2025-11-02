package main

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vaidikcode/minipay/config"
	"github.com/vaidikcode/minipay/routes"
	"github.com/vaidikcode/minipay/utils"
	"github.com/vaidikcode/minipay/workers"
)

func main() {
	utils.InitLogger()
	config.InitDB("minipay.db")

	go workers.StartWebhookWorker(1 * time.Second)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	routes.Register(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
