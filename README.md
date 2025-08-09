# Mist Prometheus Exporter

[![Build Status](https://github.com/gregwight/mistexporter/actions/workflows/release.yml/badge.svg)](https://github.com/gregwight/mistexporter/actions/workflows/release.yml)

A Prometheus exporter for Juniper Mist API metrics.

This exporter collects metrics from the Mist API, focusing on the operational state of wireless devices (APs) and their connected clients. It is designed for organizations using Prometheus and Grafana to monitor their Juniper Mist wireless environment.

## Features

-   **Hybrid Data Collection:** Uses a combination of real-time streaming and periodic scraping for efficiency and data freshness.
    -   **Real-time Streaming:** Leverages the Mist WebSocket API to stream high-frequency updates for device (AP) and client statistics.
    -   **Periodic Scraping:** Gathers organization-level metrics (e.g., alarm/ticket counts) during each Prometheus scrape.
-   **Automatic Org ID Discovery:** Automatically discovers the Organization ID from the API key if not specified, simplifying configuration.
-   **Dynamic Site Management:** Automatically discovers new sites and stops collecting metrics for sites that are removed from the organization.
-   Configurable via YAML file or environment variables.
-   Graceful shutdown and robust server lifecycle management.

## How It Works

The exporter employs a hybrid strategy to gather metrics efficiently:

1.  **Streaming (WebSocket):** On startup, the exporter establishes a WebSocket connection to the Mist API for each site in the organization. It subscribes to real-time updates for device (AP) and client statistics. These metrics are held in memory and are served instantly when Prometheus scrapes the exporter. This provides near real-time data without overwhelming the REST API.

2.  **Scraping (REST API):** For less volatile, organization-wide data (like the total number of sites or the count of open alarms), the exporter queries the Mist REST API directly at the time of the Prometheus scrape.

## Getting Started

### Prerequisites

-   A Juniper Mist API token with at least `Read-Only` access to your organization. You can generate one from the Mist Dashboard under `My Account > API Tokens`.
-   Go 1.21+ (for building from source).
-   Docker (for running as a container).

### Configuration

The exporter is configured via a `config.yaml` file.

```yaml
---
# Mist Prometheus Exporter Configuration

# The ID of the Mist Organization to scrape.
# If omitted, the exporter will attempt to discover it from the API key.
# This is required if the API key has access to multiple organizations.
# Can be set with MIST_ORG_ID environment variable.
org_id: "${MIST_ORG_ID}"

mist_api:
  # Mist API base URL.
  base_url: "https://api.mist.com"

  # Mist API token for authentication.
  # It is strongly recommended to set this via an environment variable.
  api_key: "${MIST_API_KEY}"
  
  # Timeout for individual REST API requests.
  timeout: 10s

exporter:
  # Exporter bind address.
  address: 0.0.0.0
  
  # Port on which to expose the /metrics endpoint.
  port: 10038

collector:
  # Timeout for the REST API portion of a Prometheus scrape. This should be
  # less than your Prometheus scrape_timeout setting.
  collect_timeout: 15s

  # How often to update device MAC to name mappings for the organization.
  device_name_refresh_interval: 1m

  # How often to check for new or removed sites in the organization.
  site_refresh_interval: 1m

  # Optional: Filter which sites to collect metrics from.
  # The filter will match site names using glob patterns and is case-sensitive.
  # 'include' sites with names matching the glob patterns, exlude all others.
  # 'exclude' sites with names matching the glob patterns, include all others.
  # If a site matches both include and exclude, it will be excluded.
  # See [here](https://pkg.go.dev/path/filepath#Match) for an explanation of the supported pattern syntax.
  site_filter:
    include: 
      - "US-*"
      - "EU-*"
    exclude: 
      - "*-Test"
```

### Running with Docker

A Docker image can be used to run the exporter.

1.  Create your `config.yaml`.
2.  Run the container:

```sh
docker run -d \
  --name mistexporter \
  -p 10038:10038 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  <your-docker-image-name>:latest
```

### Building from Source

1.  Clone the repository:
    ```sh
    git clone https://github.com/gregwight/mistexporter.git
    cd mistexporter
    ```
2.  Build the binary:
    ```sh
    go build -o mistexporter ./cmd/main.go
    ```
3.  Run the exporter:
    ```sh
    ./mistexporter --config config.yaml
    ```

### Running from a Pre-built Binary

Pre-built binaries for major platforms (Linux, Windows, macOS) are available on the project's GitHub Releases page. The release archives contain the exporter binary and an example `config.yaml`.

This is the recommended way to deploy the exporter without setting up a Go development environment.

**Example for Linux:**

```sh
# Replace <version> with the latest release version (e.g., v1.0.0)
VERSION="<latest_version>"
ARCH="amd64" # or arm64

# Download and extract the release archive
curl -sSL https://github.com/gregwight/mistexporter/releases/download/${VERSION}/mistexporter-${VERSION}-linux-${ARCH}.tar.gz | tar -xz

# Copy the example config and edit it with your API key and org ID
cp config.yaml.dist config.yaml
vim config.yaml # Or your favorite editor

# Run the exporter
./mistexporter --config config.yaml
```

### Running as a systemd Service (Linux)

To run the Mist exporter as a long-running service on a Linux system with `systemd`, you can create a unit file. This ensures the exporter starts on boot and is restarted if it fails.

1.  **Install the exporter:**
    Follow the steps in "Running from a Pre-built Binary" to download and extract the exporter. Then, move the files to standard locations.

    ```sh
    sudo mv mistexporter /usr/local/bin/
    sudo mkdir -p /etc/mistexporter
    sudo mv config.yaml /etc/mistexporter/
    sudo chmod +x /usr/local/bin/mistexporter
    ```

2.  **Create a dedicated user (Recommended):**
    It is a security best practice to run services as a non-root user.
    ```sh
    sudo useradd --no-create-home --shell /bin/false prometheus
    sudo chown -R prometheus:prometheus /etc/mistexporter
    ```

3.  **Create the systemd unit file:**
    Create a file at `/etc/systemd/system/mistexporter.service` with the following content:

    ```ini
    [Unit]
    Description=Mist Prometheus Exporter
    Wants=network-online.target
    After=network-online.target

    [Service]
    User=prometheus
    Group=prometheus
    Type=simple
    ExecStart=/usr/local/bin/mistexporter --config /etc/mistexporter/config.yaml
    Restart=on-failure

    [Install]
    WantedBy=multi-user.target
    ```

4.  **Enable and start the service:**
    ```sh
    # Reload the systemd daemon to recognize the new service
    sudo systemctl daemon-reload

    # Enable the service to start on boot
    sudo systemctl enable mistexporter.service

    # Start the service now
    sudo systemctl start mistexporter.service

    # Check the status
    sudo systemctl status mistexporter.service
    ```

## Exposed Metrics

The exporter exposes the following metrics at the `/metrics` endpoint.

### Scraped Metrics (On-Demand)

These metrics are fetched from the Mist REST API each time Prometheus scrapes the exporter. They are suitable for data that changes infrequently.

#### Organization Metrics
| Metric | Description | Type |
|---|---|---|
| `mist_org_alarms` | The total number of alarms in the organization. | Counter |
| `mist_org_tickets`| The total number of tickets in the organization. | Counter |

#### Site Metrics
| Metric | Description | Type |
|---|---|---|
| `mist_site_modified_time` | The last time site was modified, as a Unix timestamp. | Gauge |
| `mist_site_num_ap` | Total number of APs configured for the site. | Gauge |
| `mist_site_num_ap_connected` | Number of APs currently online at the site. | Gauge |
| `mist_site_num_clients` | Total number of clients currently connected to the site. | Gauge |
| `mist_site_num_devices` | Total number of Mist devices (APs, switches, gateways) at the site. | Gauge |
| `mist_site_num_devices_connected` | Number of Mist devices (APs, switches, gateways) currently online at the site. | Gauge |
| `mist_site_num_gateway` | Total number of gateways configured for the site. | Gauge |
| `mist_site_num_gateway_connected` | Number of gateways currently online at the site. | Gauge |
| `mist_site_num_switch` | Total number of switches configured for the site. | Gauge |
| `mist_site_num_switch_connected` | Number of switches currently online at the site. | Gauge |

### Streamed Metrics (Real-Time)

These metrics are continuously updated in the background via a WebSocket connection to the Mist API. This provides the most up-to-date information for dynamic operational data.

#### Device (AP) Metrics

All device metrics are gauges and share a common set of labels identifying the site and device (`site_name`,  `device_name`). Metrics specific to a radio also include a `radio` label (e.g., `2.4GHz`, `5GHz`).

| Metric | Description | Type |
|---|---|---|
| `mist_device_cpu_utilization_system_percent` | Current system CPU utilization of the device. | Gauge |
| `mist_device_cpu_utilization_idle_percent` | Current idle CPU utilization of the device. | Gauge |
| `mist_device_cpu_utilization_interrupt_percent` | Current interrupt CPU utilization of the device. | Gauge |
| `mist_device_cpu_utilization_user_percent` | Current user CPU utilization of the device. | Gauge |
| `mist_device_last_seen_timestamp_seconds` | The last time the device was seen, as a Unix timestamp. | Gauge |
| `mist_device_load_average_1m` | Current 1m load average of the device. | Gauge |
| `mist_device_load_average_5m` | Current 5m load average of the device. | Gauge |
| `mist_device_load_average_15m` | Current 15m load average of the device. | Gauge |
| `memory_utilization_percent` | Current memory utilization of the device. | Gauge |
| `mist_device_receive_bits_per_second` | Bits per second received by the device. | Gauge |
| `mist_device_transmit_bits_per_second` | Bits per second transmitted by the device. | Gauge |
| `mist_device_uptime_seconds` | Device uptime in seconds. | Gauge |
| `mist_device_radio_bandwidth_mhz` | Radio channel bandwidth in MHz. | Gauge |
| `mist_device_radio_channel` | The current radio channel. | Gauge |
| `mist_device_radio_clients_total` | Number of clients connected to this radio. | Gauge |
| `mist_device_radio_receive_bytes_total` | Total bytes received by the radio. | Gauge |
| `mist_device_radio_receive_packets_total` | Total packets received by the radio. | Gauge |
| `mist_device_radio_transmit_bytes_total` | Total bytes transmitted by the radio. | Gauge |
| `mist_device_radio_transmit_packets_total` | Total packets transmitted by the radio. | Gauge |
| `mist_device_radio_transmit_power_dbm` | The radio's transmit power in dBm. | Gauge |

#### Client Metrics

All client metrics are gauges and share a common set of labels identifying the site, client, and connection details (`site_name`, `device_name`, `device_mac`, `client_mac`, `client_username`, `client_hostname`, `client_os`, `client_manufacture`, `client_family`, `client_model`, `device_id`, `proto`, `radio`, `ssid`).

| Metric | Description | Type |
|---|---|---|
| `mist_client_channel` | The channel the client is connected on. | Gauge |
| `mist_client_dual_band_capable` | Whether the client is dual-band capable (1 for true, 0 for false). | Gauge |
| `mist_client_idle_seconds` | Time in seconds since the client was last active. | Gauge |
| `mist_client_is_guest_status` | Whether the client is a guest user (1 for true, 0 for false). | Gauge |
| `mist_client_last_seen_timestamp_seconds` | The last time the client was seen, as a Unix timestamp. | Gauge |
| `mist_client_locating_aps_total` | The number of APs that can hear the client. | Gauge |
| `mist_client_power_saving_mode_active` | Whether the client is in power-saving mode (1 for true, 0 for false). | Gauge |
| `mist_client_receive_bits_per_second` | Bits per second received from the client. | Gauge |
| `mist_client_receive_bytes_total` | Total bytes received from the client. | Gauge |
| `mist_client_receive_packets_total` | Total packets received from the client. | Gauge |
| `mist_client_receive_rate_mbps` | The receive data rate in Mbps. | Gauge |
| `mist_client_receive_retries_total` | Total number of receive retries. | Gauge |
| `mist_client_rssi_dbm` | The client's Received Signal Strength Indicator in dBm. | Gauge |
| `mist_client_snr_db` | The client's Signal-to-Noise Ratio in dB. | Gauge |
| `mist_client_transmit_bits_per_second` | Bits per second transmitted to the client. | Gauge |
| `mist_client_transmit_bytes_total` | Total bytes transmitted to the client. | Gauge |
| `mist_client_transmit_packets_total` | Total packets transmitted to the client. | Gauge |
| `mist_client_transmit_rate_mbps` | The transmit data rate in Mbps. | Gauge |
| `mist_client_transmit_retries_total` | Total number of transmit retries. | Gauge |
| `mist_client_uptime_seconds` | The client's session uptime in seconds. | Gauge |

## Contributing

Contributions are welcome! Please see CONTRIBUTING.md for details.

## License

This project is licensed under the Apache 2.0 License. See the LICENSE file for details.
