package utils

import (
	"log"
	"strconv"
	"sync/atomic"
)

type metricCollector struct {
	totalCharges    int64
	totalRefunds    int64
	pendingWebhooks int64
	webhookRetries  int64
	deliveredHooks  int64
	failedHooks     int64
}

var Metrics = &metricCollector{}

func (m *metricCollector) IncCharges() {
	atomic.AddInt64(&m.totalCharges, 1)
}

func (m *metricCollector) IncRefunds() {
	atomic.AddInt64(&m.totalRefunds, 1)
}

func (m *metricCollector) TotalCharges() int64 {
	return atomic.LoadInt64(&m.totalCharges)
}

func (m *metricCollector) TotalRefunds() int64 {
	return atomic.LoadInt64(&m.totalRefunds)
}

func (m *metricCollector) IncPendingWebhooks() {
	atomic.AddInt64(&m.pendingWebhooks, 1)
}

func (m *metricCollector) DecPendingWebhooks() {
	atomic.AddInt64(&m.pendingWebhooks, -1)
}

func (m *metricCollector) PendingWebhooks() int64 {
	return atomic.LoadInt64(&m.pendingWebhooks)
}

func (m *metricCollector) IncWebhookRetries() {
	atomic.AddInt64(&m.webhookRetries, 1)
}

func (m *metricCollector) WebhookRetries() int64 {
	return atomic.LoadInt64(&m.webhookRetries)
}

func (m *metricCollector) IncDeliveredHooks() {
	atomic.AddInt64(&m.deliveredHooks, 1)
}

func (m *metricCollector) DeliveredHooks() int64 {
	return atomic.LoadInt64(&m.deliveredHooks)
}

func (m *metricCollector) IncFailedHooks() {
	atomic.AddInt64(&m.failedHooks, 1)
}

func (m *metricCollector) FailedHooks() int64 {
	return atomic.LoadInt64(&m.failedHooks)
}

func InitLogger() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func Itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}
