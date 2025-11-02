package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/vaidikcode/minipay/controllers"
	"github.com/vaidikcode/minipay/utils"
)

func Register(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		api.POST("/charges", controllers.Charge)
		api.POST("/refunds", controllers.Refund)
		api.GET("/balance", controllers.Balance)
	}

	r.GET("/metrics", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"total_charges":      utils.Metrics.TotalCharges(),
			"total_refunds":      utils.Metrics.TotalRefunds(),
			"pending_webhooks":   utils.Metrics.PendingWebhooks(),
			"delivered_webhooks": utils.Metrics.DeliveredHooks(),
			"failed_webhooks":    utils.Metrics.FailedHooks(),
			"webhook_retries":    utils.Metrics.WebhookRetries(),
		})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
