package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/vaidikcode/minipay/config"
	"github.com/vaidikcode/minipay/models"
	"github.com/vaidikcode/minipay/utils"
	"github.com/vaidikcode/minipay/workers"
)

func TestWebhookRetryLogic(t *testing.T) {
	os.Remove("test_webhook.db")
	config.InitDB("test_webhook.db")

	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "temporary error")
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "success")
	}))
	defer ts.Close()

	event := models.WebhookEvent{
		TransactionID: "txn_retry_test",
		EventType:     "payment.succeeded",
		Payload:       `{"id":"txn_retry_test","amount":1000}`,
		TargetURL:     ts.URL,
		Status:        "pending",
		Attempts:      0,
		NextRunAt:     time.Now(),
	}
	config.DB.Create(&event)
	utils.Metrics.IncPendingWebhooks()

	go workers.StartWebhookWorker(100 * time.Millisecond)
	time.Sleep(4 * time.Second)

	var result models.WebhookEvent
	config.DB.First(&result, "transaction_id = ?", "txn_retry_test")

	if result.Status != "delivered" {
		t.Fatalf("expected status 'delivered', got '%s'", result.Status)
	}

	if result.Attempts < 3 {
		t.Fatalf("expected at least 3 attempts, got %d", result.Attempts)
	}
}

func TestWebhookMaxRetriesExceeded(t *testing.T) {
	os.Remove("test_webhook_maxretry.db")
	config.InitDB("test_webhook_maxretry.db")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "always fails")
	}))
	defer ts.Close()

	event := models.WebhookEvent{
		TransactionID: "txn_maxretry",
		EventType:     "payment.succeeded",
		Payload:       `{"id":"txn_maxretry","amount":2000}`,
		TargetURL:     ts.URL,
		Status:        "pending",
		Attempts:      0,
		NextRunAt:     time.Now(),
	}
	config.DB.Create(&event)
	utils.Metrics.IncPendingWebhooks()

	go workers.StartWebhookWorker(100 * time.Millisecond)
	time.Sleep(5 * time.Second)

	var result models.WebhookEvent
	config.DB.First(&result, "transaction_id = ?", "txn_maxretry")

	if result.Status != "failed" {
		t.Fatalf("expected status 'failed', got '%s'", result.Status)
	}

	if result.Attempts <= 5 {
		t.Fatalf("expected more than 5 attempts before marking failed, got %d", result.Attempts)
	}
}

func TestWebhookImmedediateSuccess(t *testing.T) {
	os.Remove("test_webhook_success.db")
	config.InitDB("test_webhook_success.db")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	event := models.WebhookEvent{
		TransactionID: "txn_immediate",
		EventType:     "payment.succeeded",
		Payload:       `{"id":"txn_immediate","amount":500}`,
		TargetURL:     ts.URL,
		Status:        "pending",
		Attempts:      0,
		NextRunAt:     time.Now(),
	}
	config.DB.Create(&event)
	utils.Metrics.IncPendingWebhooks()

	go workers.StartWebhookWorker(100 * time.Millisecond)
	time.Sleep(1 * time.Second)

	var result models.WebhookEvent
	config.DB.First(&result, "transaction_id = ?", "txn_immediate")

	if result.Status != "delivered" {
		t.Fatalf("expected status 'delivered', got '%s'", result.Status)
	}

	if result.Attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", result.Attempts)
	}
}

func TestWebhookClient4xxError(t *testing.T) {
	os.Remove("test_webhook_4xx.db")
	config.InitDB("test_webhook_4xx.db")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "bad request")
	}))
	defer ts.Close()

	event := models.WebhookEvent{
		TransactionID: "txn_4xx",
		EventType:     "payment.succeeded",
		Payload:       `{"id":"txn_4xx","amount":1500}`,
		TargetURL:     ts.URL,
		Status:        "pending",
		Attempts:      0,
		NextRunAt:     time.Now(),
	}
	config.DB.Create(&event)
	utils.Metrics.IncPendingWebhooks()

	go workers.StartWebhookWorker(100 * time.Millisecond)
	time.Sleep(1 * time.Second)

	var result models.WebhookEvent
	config.DB.First(&result, "transaction_id = ?", "txn_4xx")

	if result.Status != "failed" {
		t.Fatalf("expected status 'failed' for 4xx, got '%s'", result.Status)
	}
}

func TestRetryBackoff(t *testing.T) {
	backoff1 := utils.RetryBackoff(1)
	backoff2 := utils.RetryBackoff(2)
	backoff3 := utils.RetryBackoff(3)

	if backoff1 != time.Second {
		t.Fatalf("expected 1s backoff, got %v", backoff1)
	}

	if backoff2 != 2*time.Second {
		t.Fatalf("expected 2s backoff, got %v", backoff2)
	}

	if backoff3 != 4*time.Second {
		t.Fatalf("expected 4s backoff, got %v", backoff3)
	}

	backoff30 := utils.RetryBackoff(30)
	if backoff30 != 30*time.Second {
		t.Fatalf("expected max 30s backoff, got %v", backoff30)
	}
}
