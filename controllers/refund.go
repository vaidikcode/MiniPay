package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vaidikcode/minipay/config"
	"github.com/vaidikcode/minipay/models"
	"github.com/vaidikcode/minipay/utils"
)

type RefundRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

type RefundResponse struct {
	ID         string `json:"id"`
	Amount     int64  `json:"amount"`
	Status     string `json:"status"`
	RefundedAt string `json:"refunded_at"`
}

func Refund(c *gin.Context) {
	var req RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var txn models.Transaction
	if err := config.DB.First(&txn, "id = ?", req.TransactionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	if txn.Status != "succeeded" {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot refund transaction with status: " + txn.Status})
		return
	}

	if txn.Refunded {
		c.JSON(http.StatusConflict, gin.H{"error": "transaction already refunded"})
		return
	}

	txn.Refunded = true
	txn.Status = "refunded"
	if err := config.DB.Save(&txn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refund transaction"})
		return
	}

	utils.Metrics.IncRefunds()

	c.JSON(http.StatusOK, RefundResponse{
		ID:         txn.ID,
		Amount:     txn.Amount,
		Status:     txn.Status,
		RefundedAt: time.Now().Format(time.RFC3339),
	})
}
