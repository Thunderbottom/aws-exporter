package main

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/thunderbottom/aws-cost-exporter/exporter"
	"github.com/thunderbottom/aws-cost-exporter/config"
)

var (
	cfg         = config.GetConfig()
	logger      = getLogger()
	// AWSExporter is an instance of the Exporter structure
	AWSExporter = &exporter.Exporter{
		Config: &cfg,
		Logger: logger,
	}
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
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout * time.Millisecond,
		WriteTimeout: cfg.Server.WriteTimeout * time.Millisecond,
	}

	AWSExporter.SetAWSSession()

	logger.Infof("Starting server. Listening on: %v", cfg.Server.Address)
	if err := server.ListenAndServe(); err != nil {
		logger.Fatalf("error starting server: %v", err)
	}
}
