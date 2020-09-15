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
	if err := AWSExporter.CollectCostMetrics(); err != nil {
		w.Write([]byte("An error has occurred while collecting Cost Exporter metrics, check the logs for more information."))
		return
	}
	if err := AWSExporter.CollectInstanceMetrics(); err != nil {
		w.Write([]byte("An error has occurred while collecting EC2 Instance metrics, check the logs for more information."))
		return
	}
	AWSExporter.Metrics.WritePrometheus(w)
}
