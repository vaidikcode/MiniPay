package workers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/vaidikcode/minipay/config"
	"github.com/vaidikcode/minipay/models"
	"github.com/vaidikcode/minipay/utils"
)

func StartWebhookWorker(pollInterval time.Duration) {
	client := &http.Client{Timeout: 10 * time.Second}

	for {
		now := time.Now()
		var events []models.WebhookEvent

		if err := config.DB.Where("status = ? AND next_run_at <= ?", "pending", now).Find(&events).Error; err != nil {
			time.Sleep(pollInterval)
			continue
		}

		for _, event := range events {
			go deliverWebhook(client, event)
		}

		time.Sleep(pollInterval)
	}
}

func deliverWebhook(client *http.Client, event models.WebhookEvent) {
	var payload interface{}
	if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
		markWebhookFailed(&event)
		return
	}

	reqBody := bytes.NewBuffer([]byte(event.Payload))
	req, err := http.NewRequest("POST", event.TargetURL, reqBody)
	if err != nil {
		scheduleWebhookRetry(&event)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", event.EventType)
	req.Header.Set("X-Webhook-Delivery", utils.Itoa(int64(event.ID)))

	resp, err := client.Do(req)
	if err != nil {
		scheduleWebhookRetry(&event)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		markWebhookDelivered(&event)
		return
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		markWebhookFailed(&event)
		return
	}

	scheduleWebhookRetry(&event)
}

func markWebhookDelivered(event *models.WebhookEvent) {
	event.Status = "delivered"
	event.Attempts++
	config.DB.Save(event)
	utils.Metrics.DecPendingWebhooks()
	utils.Metrics.IncDeliveredHooks()
}

func markWebhookFailed(event *models.WebhookEvent) {
	event.Status = "failed"
	event.Attempts++
	config.DB.Save(event)
	utils.Metrics.DecPendingWebhooks()
	utils.Metrics.IncFailedHooks()
}

func scheduleWebhookRetry(event *models.WebhookEvent) {
	event.Attempts++

	if event.Attempts > 5 {
		markWebhookFailed(event)
		return
	}

	backoff := utils.RetryBackoff(event.Attempts)
	event.NextRunAt = time.Now().Add(backoff)
	event.Status = "pending"
	config.DB.Save(event)
	utils.Metrics.IncWebhookRetries()
}
