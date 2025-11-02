package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/vaidikcode/minipay/config"
	"github.com/vaidikcode/minipay/models"
	"github.com/vaidikcode/minipay/routes"
)

func setupTestDB(t *testing.T) {
	os.Remove("test_api.db")
	config.InitDB("test_api.db")
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	routes.Register(r)
	return r
}

func TestChargeCreateTransaction(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	payload := map[string]interface{}{
		"amount":   1000,
		"currency": "usd",
		"customer": "cust_123",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/charges", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem-test-1")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["id"] == nil {
		t.Fatal("expected transaction id in response")
	}
	if resp["idempotency_key"] != "idem-test-1" {
		t.Fatal("expected idempotency key in response")
	}
}

func TestChargeIdempotency(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	payload := map[string]interface{}{
		"amount":   2000,
		"currency": "usd",
		"customer": "cust_456",
	}
	body, _ := json.Marshal(payload)

	req1, _ := http.NewRequest("POST", "/api/v1/charges", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Idempotency-Key", "idem-dup-test")

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	var resp1 map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &resp1)
	txnID1 := resp1["id"]

	req2, _ := http.NewRequest("POST", "/api/v1/charges", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", "idem-dup-test")

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	var resp2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp2)
	txnID2 := resp2["id"]

	if txnID1 != txnID2 {
		t.Fatalf("expected same transaction id for duplicate idem key, got %v and %v", txnID1, txnID2)
	}
}

func TestChargeInvalidPayload(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	payload := map[string]interface{}{
		"currency": "usd",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/charges", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestRefundTransaction(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	txn := models.Transaction{
		ID:       "txn_refund_test",
		Amount:   5000,
		Currency: "usd",
		Customer: "cust_789",
		Status:   "succeeded",
	}
	config.DB.Create(&txn)

	payload := map[string]interface{}{
		"transaction_id": "txn_refund_test",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/refunds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var result models.Transaction
	config.DB.First(&result, "id = ?", "txn_refund_test")
	if !result.Refunded || result.Status != "refunded" {
		t.Fatal("expected transaction to be refunded")
	}
}

func TestRefundNonexistentTransaction(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	payload := map[string]interface{}{
		"transaction_id": "txn_nonexistent",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/refunds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestRefundAlreadyRefunded(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	txn := models.Transaction{
		ID:       "txn_already_refunded",
		Amount:   3000,
		Currency: "usd",
		Customer: "cust_999",
		Status:   "refunded",
		Refunded: true,
	}
	config.DB.Create(&txn)

	payload := map[string]interface{}{
		"transaction_id": "txn_already_refunded",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", "/api/v1/refunds", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w.Code)
	}
}

func TestGetBalance(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	config.DB.Create(&models.Transaction{
		ID:       "txn_bal_1",
		Amount:   1000,
		Currency: "usd",
		Customer: "cust_bal",
		Status:   "succeeded",
		Refunded: false,
	})

	config.DB.Create(&models.Transaction{
		ID:       "txn_bal_2",
		Amount:   500,
		Currency: "usd",
		Customer: "cust_bal",
		Status:   "succeeded",
		Refunded: true,
	})

	req, _ := http.NewRequest("GET", "/api/v1/balance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	balance := resp["balance"].(float64)
	if balance != 500 {
		t.Fatalf("expected balance 500, got %v", balance)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	req, _ := http.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["total_charges"] == nil {
		t.Fatal("expected total_charges in metrics")
	}
}

func TestHealthEndpoint(t *testing.T) {
	setupTestDB(t)
	r := setupTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
