package exporter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/sirupsen/logrus"
	"github.com/thunderbottom/aws-exporter/config"
	"golang.org/x/sync/errgroup"
)

// CostExplorer is a structure representing the functions required
// to fetch data from AWS Cost Explorer
type CostExplorer struct {
	client     *costexplorer.CostExplorer
	job        *config.Job
	logger     *logrus.Logger
	metrics    *metrics.Set
	timeperiod int
}

// CollectCostMetrics scrapes the AWS Cost Explorer API and writes the metric data to Prometheus
func (exporter *Exporter) CollectCostMetrics() error {
	ce := exporter.getCEExporter()

	var g errgroup.Group
	if len(exporter.Job.InstanceTags) > 0 {
		g.Go(ce.getCostAndUsageByTag)
	} else {
		g.Go(ce.getCostAndUsage)
	}
	g.Go(ce.getYearlyCostForecast)
	g.Go(ce.getReservationMetrics)

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (ce *CostExplorer) getCostAndUsage() error {
	costUsage, err := ce.client.GetCostAndUsage((&costexplorer.GetCostAndUsageInput{
		Metrics:     []*string{aws.String("BlendedCost")},
		TimePeriod:  getInterval(-ce.timeperiod, 0),
		Granularity: aws.String(ce.job.Granularity),
		GroupBy: []*costexplorer.GroupDefinition{
			{
				Key:  aws.String("SERVICE"),
				Type: aws.String("DIMENSION"),
			},
		},
	}))

	if err != nil {
		ce.logger.Errorf("Error occurred while retrieving cost and usage data: %s", err)
		return err
	}

	for _, cost := range costUsage.ResultsByTime[0].Groups {
		amount, err := strconv.ParseFloat(*cost.Metrics["BlendedCost"].Amount, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing cost and usage data for key %s:\n%s", *cost.Keys[0], err)
			return err
		}
		costMetric := fmt.Sprintf(`ce_cost_by_service{job="%s",service="%s"}`, ce.job.Name, *cost.Keys[0])
		ce.metrics.GetOrCreateGauge(costMetric, func() float64 {
			return amount
		})
	}

	return nil
}

func (ce *CostExplorer) getCostAndUsageByTag() error {
	for _, tag := range ce.job.InstanceTags {
		costUsage, err := ce.client.GetCostAndUsage((&costexplorer.GetCostAndUsageInput{
			Metrics:     []*string{aws.String("BlendedCost")},
			TimePeriod:  getInterval(-ce.timeperiod, 0),
			Granularity: aws.String(ce.job.Granularity),
			GroupBy: []*costexplorer.GroupDefinition{
				{
					Key:  aws.String("SERVICE"),
					Type: aws.String("DIMENSION"),
				},
				{
					Key:  aws.String(tag.Tag),
					Type: aws.String("TAG"),
				},
			},
		}))

		if err != nil {
			ce.logger.Errorf("Error occurred while retrieving cost and usage data (by tags): %s", err)
			return err
		}

		for _, cost := range costUsage.ResultsByTime[0].Groups {
			amount, err := strconv.ParseFloat(*cost.Metrics["BlendedCost"].Amount, 64)
			if err != nil {
				ce.logger.Errorf("Error occurred while parsing cost and usage data for tag %s:\n%s", *cost.Keys[1], err)
				return err
			}
			tagValue := strings.Split(*cost.Keys[1], "$")[1]
			if tagValue == "" {
				tagValue = "undefined"
			}
			tagMetric := fmt.Sprintf(`ce_cost_by_tag{job="%s",service="%s",%s="%s"}`, ce.job.Name, *cost.Keys[0], tag.ExportedTag, tagValue)
			ce.metrics.GetOrCreateGauge(tagMetric, func() float64 {
				return amount
			})
		}
	}

	return nil
}

func (ce *CostExplorer) getYearlyCostForecast() error {
	costForecast, err := ce.client.GetCostForecast((&costexplorer.GetCostForecastInput{
		Metric:      aws.String("BLENDED_COST"),
		TimePeriod:  getInterval(0, 365),
		Granularity: aws.String("MONTHLY"),
	}))

	if err != nil {
		ce.logger.Errorf("Error occurred while retrieving yearly cost forecast: %s", err)
		return err
	}

	for _, forecast := range costForecast.ForecastResultsByTime {
		amount, err := strconv.ParseFloat(*forecast.MeanValue, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing yearly cost forecast for period: %s:\n%s", *forecast.TimePeriod.Start, err)
			return err
		}
		forecastDate, err := time.Parse("2006-01-02", *forecast.TimePeriod.Start)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing forecast month: %s", err)
			return err
		}
		forecastMetric := fmt.Sprintf(`ce_forecast_by_month{job="%s",month="%s"}`, ce.job.Name, forecastDate.Month())
		ce.metrics.GetOrCreateGauge(forecastMetric, func() float64 {
			return amount
		})
	}

	yearlyForecastMetric := fmt.Sprintf(`ce_forecast_yearly{job="%s"}`, ce.job.Name)
	ce.metrics.GetOrCreateGauge(yearlyForecastMetric, func() float64 {
		total, err := strconv.ParseFloat(*costForecast.Total.Amount, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing total cost forecast:\n%s", err)
			return 0.0
		}
		return total
	})

	return nil
}

func (ce *CostExplorer) getReservationMetrics() error {
	reservationCoverage, err := ce.client.GetReservationCoverage(&costexplorer.GetReservationCoverageInput{
		Granularity: aws.String("MONTHLY"),
		TimePeriod:  getInterval(-time.Now().YearDay(), 0),
	})

	if err != nil {
		ce.logger.Errorf("Error occurred while retrieving reservation coverage: %s", err)
		return err
	}

	totalReservationHours := reservationCoverage.Total.CoverageHours

	coverageHourPerc := fmt.Sprintf(`ce_coverage_hours_percent{job="%s"}`, ce.job.Name)
	ce.metrics.GetOrCreateGauge(coverageHourPerc, func() float64 {
		total, err := strconv.ParseFloat(*totalReservationHours.CoverageHoursPercentage, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing coverage hours percentage: %s", err)
			return 0
		}
		return total
	})
	coverageOndemandHr := fmt.Sprintf(`ce_coverage_ondemand_hours{job="%s"}`, ce.job.Name)
	ce.metrics.GetOrCreateGauge(coverageOndemandHr, func() float64 {
		total, err := strconv.ParseFloat(*totalReservationHours.OnDemandHours, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing coverage ondemand hours: %s", err)
			return 0
		}
		return total
	})
	coverageReservedHr := fmt.Sprintf(`ce_coverage_reserved_hours{job="%s"}`, ce.job.Name)
	ce.metrics.GetOrCreateGauge(coverageReservedHr, func() float64 {
		total, err := strconv.ParseFloat(*totalReservationHours.ReservedHours, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing total reservation hours: %s", err)
			return 0
		}
		return total
	})
	coverageTotalHr := fmt.Sprintf(`ce_coverage_total_running_hours{job="%s"}`, ce.job.Name)
	ce.metrics.GetOrCreateGauge(coverageTotalHr, func() float64 {
		total, err := strconv.ParseFloat(*totalReservationHours.TotalRunningHours, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing coverage total hours: %s", err)
			return 0
		}
		return total
	})

	reservationUtilization, err := ce.client.GetReservationUtilization(&costexplorer.GetReservationUtilizationInput{
		Granularity: aws.String("MONTHLY"),
		TimePeriod:  getInterval(-time.Now().YearDay(), 0),
	})

	if err != nil {
		ce.logger.Errorf("Error occurred while retrieving reservation utilization: %s", err)
		return err
	}

	reservationUtil := fmt.Sprintf(`ce_reserved_utilization_percent{job="%s"}`, ce.job.Name)
	ce.metrics.GetOrCreateGauge(reservationUtil, func() float64 {
		total, err := strconv.ParseFloat(*reservationUtilization.Total.UtilizationPercentage, 64)
		if err != nil {
			ce.logger.Errorf("Error occurred while parsing reserved utilization percent: %s", err)
			return 0
		}
		return total
	})

	return nil
}

func getInterval(start int, end int) *costexplorer.DateInterval {
	dateInterval := costexplorer.DateInterval{}
	now := time.Now()
	startDate := now.AddDate(0, 0, start)
	endDate := now.AddDate(0, 0, end)

	if startDate == endDate {
		startDate = startDate.AddDate(0, 0, -1)
	}

	dateInterval.SetStart(startDate.Format("2006-01-02"))
	dateInterval.SetEnd(endDate.Format("2006-01-02"))

	return &dateInterval
}

func (exporter *Exporter) getCEExporter() *CostExplorer {
	var client *costexplorer.CostExplorer
	if exporter.Job.AWS.RoleARN != "" {
		creds := stscreds.NewCredentials(exporter.Session, exporter.Job.AWS.RoleARN)
		client = costexplorer.New(exporter.Session, &aws.Config{Credentials: creds})
	} else {
		client = costexplorer.New(exporter.Session)
	}

	ce := &CostExplorer{
		client:     client,
		job:        exporter.Job,
		logger:     exporter.Logger,
		metrics:    exporter.Metrics,
	}

	switch exporter.Job.Granularity {
	case "hourly":
		ce.job.Granularity = costexplorer.GranularityHourly
		ce.timeperiod = 1
	case "daily":
		ce.job.Granularity = costexplorer.GranularityDaily
		ce.timeperiod = 1
	case "weekly":
		ce.job.Granularity = costexplorer.GranularityDaily
		ce.timeperiod = 7
	case "monthly":
		ce.job.Granularity = costexplorer.GranularityMonthly
		ce.timeperiod = time.Now().Day() - 1
	default:
		exporter.Job.Granularity = costexplorer.GranularityDaily
		ce.timeperiod = 1
	}

	return ce
}
