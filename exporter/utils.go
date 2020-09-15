package exporter

import (
	"github.com/sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/VictoriaMetrics/metrics"
	"github.com/thunderbottom/aws-cost-exporter/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// Exporter representst the structure for all app wide exporters and structs
type Exporter struct {
	Logger      *logrus.Logger
	Config      *config.Config
	Session     *session.Session
	Metrics     *metrics.Set
}

// SetAWSSession is a method to create a new session for the AWS API
func (exporter *Exporter) SetAWSSession() {
	config := &aws.Config{
		Region: aws.String(exporter.Config.AWS.Region),
	}
	if exporter.Config.AWS.AccessKey != "" && exporter.Config.AWS.SecretKey != "" {
		config.Credentials = credentials.NewStaticCredentials(
				exporter.Config.AWS.AccessKey,
				exporter.Config.AWS.SecretKey,
				"")
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: *config,
	}))
	exporter.Session = sess
}
