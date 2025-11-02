package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vaidikcode/minipay/config"
	"github.com/vaidikcode/minipay/models"
	"github.com/vaidikcode/minipay/utils"
)

type ChargeRequest struct {
	Amount   int64  `json:"amount" binding:"required,gt=0"`
	Currency string `json:"currency" binding:"required"`
	Customer string `json:"customer" binding:"required"`
}

type ChargeResponse struct {
	ID             string `json:"id"`
	Amount         int64  `json:"amount"`
	Currency       string `json:"currency"`
	Customer       string `json:"customer"`
	Status         string `json:"status"`
	IdempotencyKey string `json:"idempotency_key"`
	CreatedAt      string `json:"created_at"`
}

func Charge(c *gin.Context) {
	var req ChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idemKey := c.GetHeader("Idempotency-Key")
	if idemKey == "" {
		idemKey = uuid.NewString()
	}

	var existing models.IdempotencyKey
	if err := config.DB.First(&existing, "id = ?", idemKey).Error; err == nil {
		var txn models.Transaction
		if err := config.DB.First(&txn, "id = ?", existing.TransactionID).Error; err == nil {
			c.JSON(http.StatusOK, ChargeResponse{
				ID:             txn.ID,
				Amount:         txn.Amount,
				Currency:       txn.Currency,
				Customer:       txn.Customer,
				Status:         txn.Status,
				IdempotencyKey: idemKey,
				CreatedAt:      txn.CreatedAt.Format(time.RFC3339),
			})
			return
		}
	}

	txnID := "txn_" + uuid.NewString()
	txn := models.Transaction{
		ID:       txnID,
		Amount:   req.Amount,
		Currency: req.Currency,
		Customer: req.Customer,
		Status:   "pending",
	}

	if err := config.DB.Create(&txn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create transaction"})
		return
	}

	idemEntry := models.IdempotencyKey{
		ID:            idemKey,
		TransactionID: txnID,
	}
	config.DB.Create(&idemEntry)

	config.DB.Model(&txn, "status = ?", "pending").Update("status", "succeeded")

	utils.Metrics.IncCharges()

	target := os.Getenv("WEBHOOK_TARGET")
	if target == "" {
		target = "http://localhost:8081/webhook"
	}

	payload := map[string]interface{}{
		"id":       txn.ID,
		"amount":   txn.Amount,
		"currency": txn.Currency,
		"customer": txn.Customer,
		"status":   "succeeded",
	}
	payloadBytes, _ := json.Marshal(payload)

	event := models.WebhookEvent{
		TransactionID: txn.ID,
		EventType:     "payment.succeeded",
		Payload:       string(payloadBytes),
		TargetURL:     target,
		Status:        "pending",
		Attempts:      0,
		NextRunAt:     time.Now(),
	}

	if err := config.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue webhook"})
		return
	}

	utils.Metrics.IncPendingWebhooks()

	c.JSON(http.StatusCreated, ChargeResponse{
		ID:             txn.ID,
		Amount:         txn.Amount,
		Currency:       txn.Currency,
		Customer:       txn.Customer,
		Status:         txn.Status,
		IdempotencyKey: idemKey,
		CreatedAt:      txn.CreatedAt.Format(time.RFC3339),
	})
}
