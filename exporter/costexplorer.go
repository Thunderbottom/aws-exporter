package exporter

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/sirupsen/logrus"
	"github.com/VictoriaMetrics/metrics"
)

// CostExplorer is a structure representing the functions required
// to fetch data from AWS Cost Explorer
type CostExplorer struct {
	client  *costexplorer.CostExplorer
	logger  *logrus.Logger
	metrics *metrics.Set
}

// CollectCostMetrics scrapes the AWS Cost Explorer API and writes the metric data to Prometheus
func (exporter *Exporter) CollectCostMetrics() {
	var client *costexplorer.CostExplorer
	if exporter.Config.AWS.RoleARN != "" {
		creds := stscreds.NewCredentials(exporter.Session, exporter.Config.AWS.RoleARN)
		client = costexplorer.New(exporter.Session, &aws.Config{Credentials: creds})
	} else {
		client = costexplorer.New(exporter.Session)
	}

	ce := &CostExplorer{
		client: client,
		logger: exporter.Logger,
		metrics: exporter.Metrics,
	}

	ce.getCostAndUsage()
	ce.getYearlyCostForecast()
	ce.getReservationMetrics()
}

func (ce *CostExplorer) getCostAndUsage() {
	costUsage, err := ce.client.GetCostAndUsage((&costexplorer.GetCostAndUsageInput{
		Metrics: []*string{aws.String("BlendedCost")},
		TimePeriod:  getInterval(-1, 0),
		Granularity: aws.String("DAILY"),
		GroupBy: []*costexplorer.GroupDefinition{
			{
				Key:  aws.String("SERVICE"),
				Type: aws.String("DIMENSION"),
			},
		},
	}))

	if err != nil {
		ce.logger.Errorf("Error occurred while retrieving cost and usage data: %s", err)
		return
	}

	for _, cost := range costUsage.ResultsByTime[0].Groups {
		amount, err := strconv.ParseFloat(*cost.Metrics["BlendedCost"].Amount, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing cost and usage data for key %s:\n%s", *cost.Keys[0], err)
			return
		}
		costMetric := fmt.Sprintf(`ce_cost_by_service{service="%s"}`, *cost.Keys[0])
		ce.metrics.GetOrCreateGauge(costMetric, func() float64 {
			return amount
		})
	}
}

func (ce *CostExplorer) getYearlyCostForecast() {
	costForecast, err := ce.client.GetCostForecast((&costexplorer.GetCostForecastInput{
		Metric: aws.String("BLENDED_COST"),
		TimePeriod: getInterval(0, 365),
		Granularity: aws.String("MONTHLY"),
	}))

	if err != nil {
		ce.logger.Errorf("Error occurred while retrieving yearly cost forecast: %s", err)
	}

	for _, forecast := range costForecast.ForecastResultsByTime {
		amount, err := strconv.ParseFloat(*forecast.MeanValue, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing yearly cost forecast for period: %s:\n%s", *forecast.TimePeriod.Start, err)
			return
		}
		forecastDate, err := time.Parse("2006-01-02", *forecast.TimePeriod.Start)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing forecast month: %s", err)
		}
		forecastMetric := fmt.Sprintf(`ce_forecast_by_month{month="%s"}`, forecastDate.Month())
		ce.metrics.GetOrCreateGauge(forecastMetric, func() float64 {
			return amount
		})
	}

	ce.metrics.GetOrCreateGauge(`ce_forecast_yearly`, func() float64 {
		total, err := strconv.ParseFloat(*costForecast.Total.Amount, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing total cost forecast:\n%s", err)
			return 0.0
		}
		return total
	})
}

func (ce *CostExplorer) getReservationMetrics() {
	reservationCoverage, err := ce.client.GetReservationCoverage(&costexplorer.GetReservationCoverageInput{
		Granularity: aws.String("MONTHLY"),
		TimePeriod: getInterval(-time.Now().YearDay(), 0),
	})

	if err != nil {
		ce.logger.Errorf("Error occurred while retrieving reservation coverage: %s", err)
		return
	}

	totalReservationHours := reservationCoverage.Total.CoverageHours

	ce.metrics.GetOrCreateGauge(`ce_coverage_hours_percent`, func() float64 {
		total, err := strconv.ParseFloat(*totalReservationHours.CoverageHoursPercentage, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing coverage hours percentage: %s", err)
			return 0
		}
		return total
	})
	ce.metrics.GetOrCreateGauge(`ce_coverage_ondemand_hours`, func() float64 {
		total, err := strconv.ParseFloat(*totalReservationHours.OnDemandHours, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing coverage ondemand hours: %s", err)
			return 0
		}
		return total
	})
	ce.metrics.GetOrCreateGauge(`ce_coverage_reserved_hours`, func() float64 {
		total, err := strconv.ParseFloat(*totalReservationHours.ReservedHours, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing total reservation hours: %s", err)
			return 0
		}
		return total
	})
	ce.metrics.GetOrCreateGauge(`ce_coverage_total_running_hours`, func() float64 {
		total, err := strconv.ParseFloat(*totalReservationHours.TotalRunningHours, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing coverage total hours: %s", err)
			return 0
		}
		return total
	})

	reservationUtilization, err := ce.client.GetReservationUtilization(&costexplorer.GetReservationUtilizationInput{
		Granularity: aws.String("MONTHLY"),
		TimePeriod: getInterval(-time.Now().YearDay(), 0),
	})

	if err != nil {
		ce.logger.Errorf("Error occurred while retrieving reservation utilization: %s", err)
		return
	}

	ce.metrics.GetOrCreateGauge(`ce_reserved_utilization_percent`, func() float64 {
		total, err := strconv.ParseFloat(*reservationUtilization.Total.UtilizationPercentage, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing reserved utilization percent: %s", err)
			return 0
		}
		return total
	})
}

func getInterval(start int, end int) *costexplorer.DateInterval {
	dateInterval := costexplorer.DateInterval{}
	now := time.Now()
	startDate := now.AddDate(0, 0, start)
	endDate := now.AddDate(0, 0, end)
	dateInterval.SetStart(startDate.Format("2006-01-02"))
	dateInterval.SetEnd(endDate.Format("2006-01-02"))

	return &dateInterval
}
