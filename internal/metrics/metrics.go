package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gregwight/mistclient"
	"github.com/gregwight/mistexporter/internal/filter"
	"github.com/prometheus/client_golang/prometheus"
)

// MistMetrics is coordinates the collection of both streamed and on-demand metrics.
type MistMetrics struct {
	client                   *mistclient.APIClient
	orgID                    string
	filter                   *filter.Filter
	siteRefreshInterval      time.Duration
	deviceNameRefresInterval time.Duration
	ready                    chan struct{}
	reg                      *prometheus.Registry
	logger                   *slog.Logger

	mu          sync.RWMutex
	sites       map[string]*StreamCollector
	deviceNames map[string]string
}

// New creates a new MistMetrics.
func New(client *mistclient.APIClient, orgID string, siteFilter *filter.Filter, siteRefreshInterval time.Duration, deviceNameRefresInterval time.Duration, reg *prometheus.Registry, logger *slog.Logger) (*MistMetrics, error) {
	if client == nil {
		return nil, fmt.Errorf("client cannot be nil")
	}

	deviceMetrics = newDeviceMetrics(reg)
	clientMetrics = newClientMetrics(reg)

	return &MistMetrics{
		client:                   client,
		orgID:                    orgID,
		filter:                   siteFilter,
		siteRefreshInterval:      siteRefreshInterval,
		deviceNameRefresInterval: deviceNameRefresInterval,
		ready:                    make(chan struct{}),
		reg:                      reg,
		logger:                   logger.With(slog.String("component", "metrics")),
		sites:                    make(map[string]*StreamCollector),
		deviceNames:              make(map[string]string),
	}, nil
}

func (c *MistMetrics) Run(ctx context.Context) error {
	wg := &sync.WaitGroup{}
	if err := c.updateDeviceNameMap(); err != nil {
		return fmt.Errorf("unable to initialize device name map: %w", err)
	}

	if err := c.manageSiteStreams(ctx, wg); err != nil {
		return fmt.Errorf("unable to initialize site metric streams: %w", err)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(c.deviceNameRefresInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.updateDeviceNameMap(); err != nil {
					c.logger.Error("unable to refresh org device names", "error", err)
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(c.siteRefreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.manageSiteStreams(ctx, wg); err != nil {
					c.logger.Error("unable to refresh site metric streams", "error", err)
				}
			}
		}
	}()

	close(c.ready)
	wg.Wait()
	return nil
}

func (c *MistMetrics) updateDeviceNameMap() error {
	c.logger.Debug("running org device name map updater...")
	defer c.logger.Debug("org device name map updater finished")

	c.mu.Lock()
	defer c.mu.Unlock()

	devices, err := c.client.ListOrgDevices(c.orgID)
	if err != nil {
		return fmt.Errorf("unable to fetch device list: %w", err)
	}

	c.deviceNames = devices
	return nil
}

func (c *MistMetrics) manageSiteStreams(ctx context.Context, wg *sync.WaitGroup) error {
	c.logger.Debug("running site metric stream manager...")
	defer c.logger.Debug("site metric stream manager finished")

	c.mu.Lock()
	defer c.mu.Unlock()

	sites, err := c.client.GetOrgSites(c.orgID)
	if err != nil {
		return fmt.Errorf("unable to fetch site list: %w", err)
	}

	activeSites := make(map[string]struct{})
	for _, site := range sites {
		if isFiltered, err := c.filter.IsFiltered(site); err != nil {
			c.logger.Error("unable to apply site filter to site", "site", site.Name, "error", err)
			continue
		} else if isFiltered {
			continue
		}

		activeSites[site.ID] = struct{}{}
		streamer, ok := c.sites[site.ID]
		if !ok {
			streamer = newStreamCollector(c.client, site, c.deviceNames, c.logger)
			c.sites[site.ID] = streamer
		}

		streamer.mu.RLock()
		if !streamer.running {
			wg.Add(1)
			go streamer.run(ctx, wg)
		}
		streamer.mu.RUnlock()
	}

	for siteID, streamer := range c.sites {
		if _, ok := activeSites[siteID]; !ok {
			streamer.stop()
			delete(c.sites, siteID)
		}
	}

	return nil
}

func (c *MistMetrics) Ready() <-chan struct{} {
	return c.ready
}

// StreamCollector coordinates the collection of metrics from a set of websocket streams.
type StreamCollector struct {
	client      *mistclient.APIClient
	site        mistclient.Site
	deviceNames map[string]string
	logger      *slog.Logger

	mu      sync.RWMutex
	running bool
	cancel  context.CancelFunc
}

func (c *StreamCollector) stop() {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cancel != nil {
		c.cancel()
	}
}

func newStreamCollector(client *mistclient.APIClient, site mistclient.Site, deviceNames map[string]string, logger *slog.Logger) *StreamCollector {
	return &StreamCollector{
		client:      client,
		site:        site,
		deviceNames: deviceNames,
		logger:      logger.With(slog.String("site", site.Name)),
	}
}

func (c *StreamCollector) run(ctx context.Context, wg *sync.WaitGroup) {
	runCtx, cancel := context.WithCancel(ctx)

	c.mu.Lock()
	c.running = true
	c.cancel = cancel
	c.mu.Unlock()

	c.logger.Info("starting site metrics stream...")
	defer func() {
		cancel()
		c.mu.Lock()
		c.running = false
		c.cancel = nil
		c.mu.Unlock()

		c.logger.Info("site metrics stream stopped")
		wg.Done()
	}()

	deviceStats, err := c.client.StreamSiteDeviceStats(runCtx, c.site.ID)
	if err != nil {
		c.logger.Error("unable to start site device stats stream", "error", err)
		return
	}
	c.logger.Debug("site device stats stream started")

	clientStats, err := c.client.StreamSiteClientStats(runCtx, c.site.ID)
	if err != nil {
		c.logger.Error("unable to start site client stats stream", "error", err)
		return
	}
	c.logger.Debug("site client stats stream started")

	// WaitGroup to ensure all subscriptions are closed before we exit.
	// If we get a failure on one channel we cancel the context to
	// force the other channels to disconnect. We will be restarted by
	// the stream manager unless the parent context is done.
	hwg := &sync.WaitGroup{}
	hwg.Add(1)
	go func() {
		defer hwg.Done()
		defer cancel()

		for stat := range deviceStats {
			// There's nothing we can do if there's no name so no need to
			// check if the key exists - missing macs will get an empty label.
			c.mu.RLock()
			deviceName := c.deviceNames[stat.Mac]
			c.mu.RUnlock()

			handleSiteDeviceStat(c.site, deviceName, stat)
		}
	}()

	hwg.Add(1)
	go func() {
		defer hwg.Done()
		defer cancel()

		for stat := range clientStats {
			// There's nothing we can do if there's no name so no need to
			// check if the key exists - missing macs will get an empty label.
			c.mu.RLock()
			deviceName := c.deviceNames[stat.APMac]
			c.mu.RUnlock()
			handleSiteClientStat(c.site, deviceName, stat)
		}
	}()

	hwg.Wait()
}

func boolToFloat64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
