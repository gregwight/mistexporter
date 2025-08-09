package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gregwight/mistclient"
	"github.com/gregwight/mistexporter/internal/collector"
	"github.com/gregwight/mistexporter/internal/config"
	"github.com/gregwight/mistexporter/internal/filter"
	"github.com/gregwight/mistexporter/internal/metrics"
	"github.com/gregwight/mistexporter/internal/server"
	"github.com/gregwight/mistexporter/internal/version"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"golang.org/x/sync/errgroup"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Path to the configuration file")
	debug := flag.Bool("debug", false, "Enable debug mode")
	version.AddVersionFlag()
	flag.Parse()

	// Create context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGKILL,
	)
	defer cancel()

	// Initialize logger
	loggerOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if *debug {
		loggerOpts.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, loggerOpts))

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		logger.Error("unable to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize Mist API client
	client, err := mistclient.New(cfg.MistClient, logger)
	if err != nil {
		logger.Error("unable to initialize Mist API client", "error", err)
		os.Exit(1)
	}

	// Initialize site filter
	siteFilter, err := filter.New(cfg.Collector.SiteFilter)
	if err != nil {
		logger.Error("unable to initialize site filter", "error", err)
		os.Exit(1)
	}

	// Determine Mist OrgID
	orgID, err := autoOrgID(cfg, client)
	if err != nil {
		logger.Error("unable to determine Mist OrgID", "error", err)
		os.Exit(1)
	}

	// Create a pedantic reg
	reg := prometheus.NewPedanticRegistry()

	// Add Go runtime metrics
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Add the scrape-time collector
	c, err := collector.New(client, orgID, siteFilter, logger)
	if err != nil {
		logger.Error("unable to initialize scrape-time collector", "error", err)
		os.Exit(1)
	}
	reg.MustRegister(c)

	// Use errgroup for managing goroutines
	eg, ctx := errgroup.WithContext(ctx)

	// Create and start metrics streamer
	m, err := metrics.New(client, orgID, siteFilter, cfg.Collector.SiteRefreshInterval, cfg.Collector.DeviceNameRefreshInterval, reg, logger)
	if err != nil {
		logger.Error("unable to initialize metrics streamer", "error", err)
		os.Exit(1)
	}
	eg.Go(func() error {
		logger.Info("starting metrics streamer...", "org_id", orgID)
		return m.Run(ctx)
	})

	select {
	case <-ctx.Done():
		if err := eg.Wait(); err != nil {
			logger.Error("metrics streamer failed to start", "error", err)
			os.Exit(1)
		}
		logger.Info("server startup terminated")
		return
	case <-m.Ready():
		logger.Info("metrics streamer started successfully")
	}

	// Create and start HTTP server
	svr, err := server.New(cfg, reg)
	if err != nil {
		logger.Error("unable to create HTTP server", "error", err)
		os.Exit(1)
	}

	eg.Go(func() error {
		logger.Info("starting HTTP server...", "address", svr.Addr)
		if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		return nil
	})

	// Handle graceful shutdown
	eg.Go(func() error {
		<-ctx.Done()
		logger.Info("shutting down HTTP server...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		return svr.Shutdown(shutdownCtx)
	})

	// Wait for all goroutines to complete
	if err := eg.Wait(); err != nil {
		logger.Error("server shutdown failure", "error", err)
		os.Exit(1)
	}

	logger.Info("server shutdown success")
}

func autoOrgID(cfg *config.Config, client *mistclient.APIClient) (string, error) {
	if cfg.OrgId != "" {
		return cfg.OrgId, nil
	}

	self, err := client.GetSelf()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve self: %w", err)
	}

	var orgID string
	for _, priv := range self.Privileges {
		if priv.Scope != "org" {
			continue
		}
		if orgID != "" {
			return "", fmt.Errorf("api key has access to multiple Mist organizations - please specify desired orgID using 'org_id' configurtaion key")
		}
		orgID = priv.OrgID
	}

	return orgID, nil
}
