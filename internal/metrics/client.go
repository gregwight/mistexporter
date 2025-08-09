package metrics

import (
	"github.com/gregwight/mistclient"
	"github.com/prometheus/client_golang/prometheus"
)

// StreamedClientLabelNames defines the labels attached to streamed wireless client metrics.
var StreamedClientLabelNames = append(SiteLabelNames,
	"device_name",
	"device_mac",
	"client_mac",
	"client_username",
	"client_hostname",
	"client_os",
	"client_manufacture",
	"client_family",
	"client_model",
	"proto",
	"radio",
	"ssid",
)

// StreamedClientLabelValues generates label values for streamed wireless client metrics.
func StreamedClientLabelValues(s mistclient.Site, deviceName string, c mistclient.StreamedClientStat) []string {
	return append(SiteLabelValues(s),
		deviceName,
		c.APMac,
		c.Mac,
		c.Username,
		c.Hostname,
		c.OS,
		c.Manufacture,
		c.Family,
		c.Model,
		c.Proto.String(),
		c.Band.String(),
		c.SSID,
	)
}

var clientMetrics *ClientMetrics

// ClientMetrics holds metrics related to wireless clients.
type ClientMetrics struct {
	channel               *prometheus.GaugeVec
	dualBandCapable       *prometheus.GaugeVec
	idleSeconds           *prometheus.GaugeVec
	isGuest               *prometheus.GaugeVec
	lastSeenTimestamp     *prometheus.GaugeVec
	locatingAps           *prometheus.GaugeVec
	powerSavingModeActive *prometheus.GaugeVec
	rssiDbm               *prometheus.GaugeVec
	receiveBps            *prometheus.GaugeVec
	receiveBytesTotal     *prometheus.GaugeVec
	receivePacketsTotal   *prometheus.GaugeVec
	receiveRateMbps       *prometheus.GaugeVec
	receiveRetriesTotal   *prometheus.GaugeVec
	snrDb                 *prometheus.GaugeVec
	transmitBps           *prometheus.GaugeVec
	transmitBytesTotal    *prometheus.GaugeVec
	transmitPacketsTotal  *prometheus.GaugeVec
	transmitRateMbps      *prometheus.GaugeVec
	transmitRetriesTotal  *prometheus.GaugeVec
	uptimeSeconds         *prometheus.GaugeVec
}

func newClientMetrics(reg *prometheus.Registry) *ClientMetrics {
	m := &ClientMetrics{
		channel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "channel",
				Help:      "The channel the client is connected on.",
			}, StreamedClientLabelNames,
		),
		dualBandCapable: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "dual_band_capable",
				Help:      "Whether the client is dual-band capable (1 for true, 0 for false).",
			}, StreamedClientLabelNames,
		),
		idleSeconds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "idle_seconds",
				Help:      "Time in seconds since the client was last active.",
			}, StreamedClientLabelNames,
		),
		isGuest: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "is_guest_status",
				Help:      "Whether the client is a guest user (1 for true, 0 for false).",
			}, StreamedClientLabelNames,
		),
		lastSeenTimestamp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "last_seen_timestamp_seconds",
				Help:      "The last time the client was seen, as a Unix timestamp.",
			}, StreamedClientLabelNames,
		),
		locatingAps: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "locating_aps_total",
				Help:      "The number of APs that can hear the client.",
			}, StreamedClientLabelNames,
		),
		powerSavingModeActive: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "power_saving_mode_active",
				Help:      "Whether the client is in power-saving mode (1 for true, 0 for false).",
			}, StreamedClientLabelNames,
		),
		rssiDbm: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "rssi_dbm",
				Help:      "The client's Received Signal Strength Indicator in dBm.",
			}, StreamedClientLabelNames,
		),
		receiveBps: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "receive_bits_per_second",
				Help:      "Bits per second received from the client.",
			}, StreamedClientLabelNames,
		),
		receiveBytesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "receive_bytes_total",
				Help:      "Total bytes received from the client.",
			}, StreamedClientLabelNames,
		),
		receivePacketsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "receive_packets_total",
				Help:      "Total packets received from the client.",
			}, StreamedClientLabelNames,
		),
		receiveRateMbps: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "receive_rate_mbps",
				Help:      "The receive data rate in Mbps.",
			}, StreamedClientLabelNames,
		),
		receiveRetriesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "receive_retries_total",
				Help:      "Total number of receive retries.",
			}, StreamedClientLabelNames,
		),
		snrDb: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "snr_db",
				Help:      "The client's Signal-to-Noise Ratio in dB.",
			}, StreamedClientLabelNames,
		),
		transmitBps: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "transmit_bits_per_second",
				Help:      "Bits per second transmitted to the client.",
			}, StreamedClientLabelNames,
		),
		transmitBytesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "transmit_bytes_total",
				Help:      "Total bytes transmitted to the client.",
			}, StreamedClientLabelNames,
		),
		transmitPacketsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "transmit_packets_total",
				Help:      "Total packets transmitted to the client.",
			}, StreamedClientLabelNames,
		),
		transmitRateMbps: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "transmit_rate_mbps",
				Help:      "The transmit data rate in Mbps.",
			}, StreamedClientLabelNames,
		),
		transmitRetriesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "transmit_retries_total",
				Help:      "Total number of transmit retries.",
			}, StreamedClientLabelNames,
		),
		uptimeSeconds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "client",
				Name:      "uptime_seconds",
				Help:      "The client's session uptime in seconds.",
			}, StreamedClientLabelNames,
		),
	}

	reg.MustRegister(
		m.channel,
		m.dualBandCapable,
		m.idleSeconds,
		m.isGuest,
		m.lastSeenTimestamp,
		m.locatingAps,
		m.powerSavingModeActive,
		m.rssiDbm,
		m.receiveBps,
		m.receiveBytesTotal,
		m.receivePacketsTotal,
		m.receiveRateMbps,
		m.receiveRetriesTotal,
		m.snrDb,
		m.transmitBps,
		m.transmitBytesTotal,
		m.transmitPacketsTotal,
		m.transmitRateMbps,
		m.transmitRetriesTotal,
		m.uptimeSeconds,
	)

	return m
}

func handleSiteClientStat(site mistclient.Site, deviceName string, stat mistclient.StreamedClientStat) {
	labels := StreamedClientLabelValues(site, deviceName, stat)

	clientMetrics.channel.WithLabelValues(labels...).Set(float64(stat.Channel))
	clientMetrics.dualBandCapable.WithLabelValues(labels...).Set(boolToFloat64(stat.DualBand))
	clientMetrics.idleSeconds.WithLabelValues(labels...).Set(stat.Idletime.Seconds())
	clientMetrics.isGuest.WithLabelValues(labels...).Set(boolToFloat64(stat.IsGuest))
	clientMetrics.lastSeenTimestamp.WithLabelValues(labels...).Set(float64(stat.LastSeen.Unix()))
	clientMetrics.locatingAps.WithLabelValues(labels...).Set(float64(stat.NumLocatingAPs))
	clientMetrics.powerSavingModeActive.WithLabelValues(labels...).Set(boolToFloat64(stat.PowerSaving))
	clientMetrics.rssiDbm.WithLabelValues(labels...).Set(float64(stat.RSSI))
	clientMetrics.receiveBps.WithLabelValues(labels...).Set(float64(stat.RxBps))
	clientMetrics.receiveBytesTotal.WithLabelValues(labels...).Set(float64(stat.RxBytes))
	clientMetrics.receivePacketsTotal.WithLabelValues(labels...).Set(float64(stat.RxPackets))
	clientMetrics.receiveRateMbps.WithLabelValues(labels...).Set(float64(stat.RxRate))
	clientMetrics.receiveRetriesTotal.WithLabelValues(labels...).Set(float64(stat.RxRetries))
	clientMetrics.snrDb.WithLabelValues(labels...).Set(float64(stat.SNR))
	clientMetrics.transmitBps.WithLabelValues(labels...).Set(float64(stat.TxBps))
	clientMetrics.transmitBytesTotal.WithLabelValues(labels...).Set(float64(stat.TxBytes))
	clientMetrics.transmitPacketsTotal.WithLabelValues(labels...).Set(float64(stat.TxPackets))
	clientMetrics.transmitRateMbps.WithLabelValues(labels...).Set(float64(stat.TxRate))
	clientMetrics.transmitRetriesTotal.WithLabelValues(labels...).Set(float64(stat.TxRetries))
	clientMetrics.uptimeSeconds.WithLabelValues(labels...).Set(stat.Uptime.Seconds())
}
