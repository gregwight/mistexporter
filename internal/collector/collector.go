package collector

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/gregwight/mistclient"
	"github.com/gregwight/mistexporter/internal/filter"
	"github.com/prometheus/client_golang/prometheus"
)

// MistCollector implements the prometheus.Collector interface.
type MistCollector struct {
	client            *mistclient.APIClient
	orgID             string
	filter            *filter.Filter
	enableOrgMetrics  bool
	enableSiteMetrics bool
	wg                *sync.WaitGroup
	logger            *slog.Logger
}

// New creates a new MistCollector.
func New(client *mistclient.APIClient, orgID string, siteFilter *filter.Filter, enableOrgMetrics bool, enableSiteMetrics bool, logger *slog.Logger) (*MistCollector, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}

	return &MistCollector{
		client:            client,
		orgID:             orgID,
		filter:            siteFilter,
		enableOrgMetrics:  enableOrgMetrics,
		enableSiteMetrics: enableSiteMetrics,
		wg:                &sync.WaitGroup{},
		logger:            logger.With(slog.String("component", "collector")),
	}, nil
}

// Describe implements the prometheus.Collector interface.
func (c *MistCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect implements the prometheus.Collector interface.
func (c *MistCollector) Collect(ch chan<- prometheus.Metric) {
	if c.enableOrgMetrics {
		// Get alarms for the organization
		c.wg.Add(1)
		go c.collectOrgAlarms(ch)

		// Get tickets for the organization
		c.wg.Add(1)
		go c.collectOrgTickets(ch)
	}

	if c.enableSiteMetrics {
		// Get sites for the organization
		c.wg.Add(1)
		go c.collectSiteStats(ch)
	}

	c.wg.Wait()
}

func (c *MistCollector) sendMetric(ch chan<- prometheus.Metric, desc *prometheus.Desc, valueType prometheus.ValueType, value float64, labels ...string) {
	ch <- prometheus.MustNewConstMetric(desc, valueType, value, labels...)
}
