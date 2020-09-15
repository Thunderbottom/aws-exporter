package main

import (
	"net/http"

	"github.com/VictoriaMetrics/metrics"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to prometheus cost exporter. Visit /metrics."))
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	AWSExporter.Metrics = metrics.NewSet()
	AWSExporter.CollectCostMetrics()
	AWSExporter.CollectInstanceMetrics()
	AWSExporter.Metrics.WritePrometheus(w)
}
