package exporter

import (
	"fmt"
	// "strconv"
	// "time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sirupsen/logrus"
	// "github.com/thunderbottom/aws-cost-exporter/config"
	"github.com/VictoriaMetrics/metrics"
)

var (
	// cfg = config.GetConfig()
	ec2Opts = session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(
				cfg.AWS.AccessKey,
				cfg.AWS.SecretKey,
				""),
			Region: aws.String(cfg.AWS.Region),
	}}
	ec2Sess = session.Must(session.NewSessionWithOptions(ec2Opts))
	ec2Service = ec2.New(ec2Sess)
)

// CollectInstanceMetrics scrapes the AWS EC2 API for Instance details and writes the metric data to Prometheus
func CollectInstanceMetrics(m *metrics.Set, logger *logrus.Logger) {
	getInstanceUsage(m, logger)
}

func getInstanceUsage(m *metrics.Set, logger *logrus.Logger) {
	var totalCoreCount int64

	instances, err := ec2Service.DescribeInstances((&ec2.DescribeInstancesInput{}))
	if err != nil {
		logger.Errorf("Error occurred while retreiving instance data: %s", err)
		return
	}

	for _, reservation := range instances.Reservations {
		for _, instance := range reservation.Instances {
			if *instance.State.Name == "running" {
				totalCoreCount = totalCoreCount + *instance.CpuOptions.CoreCount
			}
			instanceTypeMetric := fmt.Sprintf(`ce_instance_count{region="%s",type="%s",status="%s"}`, cfg.AWS.Region, *instance.InstanceType, *instance.State.Name)
			m.GetOrCreateCounter(instanceTypeMetric).Add(1)
		}
	}

	totalCoreMetric := fmt.Sprintf(`ce_total_core_count{region="%s"}`, cfg.AWS.Region)
	m.GetOrCreateGauge(totalCoreMetric, func() float64 {
		return float64(totalCoreCount)
	})
}
