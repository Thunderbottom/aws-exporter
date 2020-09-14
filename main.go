package main

import (
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/thunderbottom/aws-cost-exporter/config"
)

var (
	cfg = config.GetConfig()
	logger = getLogger()
)

func getLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	return logger
}

func main() {
	router := http.NewServeMux()
	router.Handle("/", http.HandlerFunc(defaultHandler))
	router.Handle("/metrics", http.HandlerFunc(metricsHandler))

	server := &http.Server{
		Addr: cfg.Server.Address,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		logger.Fatalf("error starting server: %v", err)
	}
}
