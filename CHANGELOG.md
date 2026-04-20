# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

### Changed

### Fixed

## [1.2.0] - 2026-04-20

### Added
- Collector-level configuration flags to control metric families:
  - enable_org_metrics (scrape-time REST org metrics)
  - enable_site_metrics (scrape-time REST site metrics)
  - enable_client_metrics (streamed websocket client metrics)

## [1.1.0] - 2025-08-08

### Added
- Support for configuring the device name refresh interval, allowing periodic updates of device names in metrics.
- Device names are now displayed in client and device metrics for improved clarity.
- Added a new commented configuration option for device name refresh interval in the sample configuration file.
- Updated README to reflect changes in metrics and configuration options.

### Fixed
- Removed the "site_id" label from all site-related metrics to streamline metric labeling.

### Changed
- Simplified device metrics to focus on CPU, load, memory, and essential radio statistics.
- Updated metric label structures for devices and clients to include device name and MAC address, and remove less relevant labels.
- Introduced periodic refreshing of device name mappings used in metrics.
- Updated tests to align with the new metric labels and configuration defaults.

## [1.0.0] - 2025-08-07

This is the initial public release of the Mist Exporter.

### Added
- Initial implementation of the Prometheus exporter for Juniper Mist.
- Hybrid data collection model using both REST API and WebSockets.
- Real-time streaming of device (AP) and client statistics for low-latency metrics.
- On-demand scraping of organization-level metrics (alarms, tickets, sites).
- Automatic discovery of Organization ID from the API key.
- Dynamic site management: automatically starts/stops metric collection for added/removed sites.
- Configuration via YAML file and environment variables.
- Graceful shutdown handling for robust operation.
- Dockerfile for containerized deployment and a `systemd` service example.
- GitHub Actions workflow for building binaries and creating releases with GoReleaser.