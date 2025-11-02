package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vaidikcode/minipay/config"
	"github.com/vaidikcode/minipay/models"
)

type BalanceResponse struct {
	SuccessfulTransactions int64 `json:"successful_transactions"`
	RefundedTransactions   int64 `json:"refunded_transactions"`
	Balance                int64 `json:"balance"`
}

func Balance(c *gin.Context) {
	var transactions []models.Transaction
	if err := config.DB.Where("status = ?", "succeeded").Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch balance"})
		return
	}

	var successful int64
	var refunded int64
	balance := int64(0)

	for _, t := range transactions {
		if t.Refunded {
			refunded++
			balance -= t.Amount
		} else {
			successful++
			balance += t.Amount
		}
	}

	c.JSON(http.StatusOK, BalanceResponse{
		SuccessfulTransactions: successful,
		RefundedTransactions:   refunded,
		Balance:                balance,
	})
}
