package main

import (
	"net/http"

	"github.com/thunderbottom/aws-cost-exporter/exporter"
	"github.com/VictoriaMetrics/metrics"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to prometheus cost exporter. Visit /metrics."))
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	m := metrics.NewSet()
	exporter.CollectCostMetrics(m, logger)
	exporter.CollectInstanceMetrics(m, logger)
	m.WritePrometheus(w)
}
