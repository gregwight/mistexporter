package metrics

import (
	"github.com/gregwight/mistclient"
	"github.com/prometheus/client_golang/prometheus"
)

// DeviceLabelNames defines the labels attached to device metrics.
var DeviceLabelNames = append(SiteLabelNames,
	"device_name",
	"device_mac",
	"device_model",
	"device_hw_rev",
)

// StreamedDeviceLabelNames defines the labels attached to streamed device metrics.
var StreamedDeviceLabelNames = append(SiteLabelNames,
	"device_name",
	"device_mac",
	"device_version",
)

// StreamedDeviceWithRadioLabelNames defines the labels attached to radio-specific device metrics.
var StreamedDeviceWithRadioLabelNames = append(StreamedDeviceLabelNames, "radio")

// DeviceLabelValues generates label values for device metrics.
func DeviceLabelValues(s mistclient.Site, d mistclient.Device) []string {
	return append(SiteLabelValues(s),
		d.Name,
		d.Mac,
		d.Model,
		d.HwRev,
	)
}

// StreamedDeviceLabelValues generates label values for streamed device metrics.
func StreamedDeviceLabelValues(s mistclient.Site, deviceName string, ds mistclient.StreamedDeviceStat) []string {
	return append(SiteLabelValues(s), deviceName, ds.Mac, ds.Version)
}

// DeviceWithRadioLabelValues generates label values for radio-specific device metrics.
func DeviceWithRadioLabelValues(s mistclient.Site, deviceName string, ds mistclient.StreamedDeviceStat, radio string) []string {
	return append(StreamedDeviceLabelValues(s, deviceName, ds), radio)
}

var deviceMetrics *DeviceMetrics

// DeviceMetrics holds metrics related to devices.
type DeviceMetrics struct {
	cpuUtilizationSystem    *prometheus.GaugeVec
	cpuUtilizationIdle      *prometheus.GaugeVec
	cpuUtilizationInterrupt *prometheus.GaugeVec
	cpuUtilizationUser      *prometheus.GaugeVec
	lastSeenTimestamp       *prometheus.GaugeVec
	loadAverage1m           *prometheus.GaugeVec
	loadAverage5m           *prometheus.GaugeVec
	loadAverage15m          *prometheus.GaugeVec
	memoryUtilization       *prometheus.GaugeVec
	receiveBps              *prometheus.GaugeVec
	transmitBps             *prometheus.GaugeVec
	uptimeSeconds           *prometheus.GaugeVec

	// Radio metrics
	radioBandwidthMhz         *prometheus.GaugeVec
	radioChannel              *prometheus.GaugeVec
	radioClients              *prometheus.GaugeVec
	radioTransmitPowerDbm     *prometheus.GaugeVec
	radioReceiveBytesTotal    *prometheus.GaugeVec
	radioReceivePacketsTotal  *prometheus.GaugeVec
	radioTransmitBytesTotal   *prometheus.GaugeVec
	radioTransmitPacketsTotal *prometheus.GaugeVec
}

func newDeviceMetrics(reg *prometheus.Registry) *DeviceMetrics {
	m := &DeviceMetrics{
		cpuUtilizationSystem: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "cpu_utilization_system_percent",
				Help:      "Current system CPU utilization of the device.",
			}, StreamedDeviceLabelNames,
		),
		cpuUtilizationIdle: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "cpu_utilization_idle_percent",
				Help:      "Current idle CPU utilization of the device.",
			}, StreamedDeviceLabelNames,
		),
		cpuUtilizationInterrupt: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "cpu_utilization_interrupt_percent",
				Help:      "Current interrupt CPU utilization of the device.",
			}, StreamedDeviceLabelNames,
		),
		cpuUtilizationUser: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "cpu_utilization_user_percent",
				Help:      "Current user CPU utilization of the device.",
			}, StreamedDeviceLabelNames,
		),
		lastSeenTimestamp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "last_seen_timestamp_seconds",
				Help:      "The last time the device was seen, as a Unix timestamp.",
			}, StreamedDeviceLabelNames,
		),
		loadAverage1m: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "load_average_1m",
				Help:      "Current 1m load average of the device.",
			}, StreamedDeviceLabelNames,
		),
		loadAverage5m: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "load_average_5m",
				Help:      "Current 5m load average of the device.",
			}, StreamedDeviceLabelNames,
		),
		loadAverage15m: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "load_average_15m",
				Help:      "Current 15m load average of the device.",
			}, StreamedDeviceLabelNames,
		),
		memoryUtilization: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "memory_utilization_percent",
				Help:      "Current memory utilization of the device.",
			}, StreamedDeviceLabelNames,
		),
		receiveBps: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "receive_bits_per_second",
				Help:      "Bits per second received by the device.",
			}, StreamedDeviceLabelNames,
		),
		transmitBps: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "transmit_bits_per_second",
				Help:      "Bits per second transmitted by the device.",
			}, StreamedDeviceLabelNames,
		),
		uptimeSeconds: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "uptime_seconds",
				Help:      "Device uptime in seconds.",
			}, StreamedDeviceLabelNames,
		),

		// Radio metrics
		radioBandwidthMhz: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "radio_bandwidth_mhz",
				Help:      "Radio channel bandwidth in MHz.",
			}, StreamedDeviceWithRadioLabelNames,
		),
		radioChannel: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "radio_channel",
				Help:      "The current radio channel.",
			}, StreamedDeviceWithRadioLabelNames,
		),
		radioClients: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "radio_clients_total",
				Help:      "Number of clients connected to this radio.",
			}, StreamedDeviceWithRadioLabelNames,
		),
		radioTransmitPowerDbm: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "radio_transmit_power_dbm",
				Help:      "The radio's transmit power in dBm.",
			}, StreamedDeviceWithRadioLabelNames,
		),
		radioReceiveBytesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "radio_receive_bytes_total",
				Help:      "Total bytes received by the radio.",
			}, StreamedDeviceWithRadioLabelNames,
		),
		radioReceivePacketsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "radio_receive_packets_total",
				Help:      "Total packets received by the radio.",
			}, StreamedDeviceWithRadioLabelNames,
		),
		radioTransmitBytesTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "radio_transmit_bytes_total",
				Help:      "Total bytes transmitted by the radio.",
			}, StreamedDeviceWithRadioLabelNames,
		),
		radioTransmitPacketsTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "mist",
				Subsystem: "device",
				Name:      "radio_transmit_packets_total",
				Help:      "Total packets transmitted by the radio.",
			}, StreamedDeviceWithRadioLabelNames,
		),
	}

	reg.MustRegister(
		m.radioBandwidthMhz,
		m.radioChannel,
		m.cpuUtilizationSystem,
		m.cpuUtilizationIdle,
		m.cpuUtilizationInterrupt,
		m.cpuUtilizationUser,
		m.lastSeenTimestamp,
		m.loadAverage1m,
		m.loadAverage5m,
		m.loadAverage15m,
		m.memoryUtilization,
		m.radioClients,
		m.radioTransmitPowerDbm,
		m.receiveBps,
		m.radioReceiveBytesTotal,
		m.radioReceivePacketsTotal,
		m.transmitBps,
		m.radioTransmitBytesTotal,
		m.radioTransmitPacketsTotal,
		m.uptimeSeconds,
	)

	return m
}

func handleSiteDeviceStat(site mistclient.Site, deviceName string, stat mistclient.StreamedDeviceStat) {
	labels := StreamedDeviceLabelValues(site, deviceName, stat)

	deviceMetrics.cpuUtilizationSystem.WithLabelValues(labels...).Set(float64(stat.CpuStat.System))
	deviceMetrics.cpuUtilizationIdle.WithLabelValues(labels...).Set(float64(stat.CpuStat.Idle))
	deviceMetrics.cpuUtilizationInterrupt.WithLabelValues(labels...).Set(float64(stat.CpuStat.Interrupt))
	deviceMetrics.cpuUtilizationUser.WithLabelValues(labels...).Set(float64(stat.CpuStat.User))
	deviceMetrics.lastSeenTimestamp.WithLabelValues(labels...).Set(float64(stat.LastSeen.Unix()))
	deviceMetrics.receiveBps.WithLabelValues(labels...).Set(float64(stat.RxBps))
	deviceMetrics.transmitBps.WithLabelValues(labels...).Set(float64(stat.TxBps))
	deviceMetrics.uptimeSeconds.WithLabelValues(labels...).Set(stat.Uptime.Seconds())

	// Radio metrics
	for radioConfig, radioStat := range stat.RadioStats {
		labels := DeviceWithRadioLabelValues(site, deviceName, stat, radioConfig.String())

		deviceMetrics.radioBandwidthMhz.WithLabelValues(labels...).Set(float64(radioStat.Bandwidth))
		deviceMetrics.radioChannel.WithLabelValues(labels...).Set(float64(radioStat.Channel))
		deviceMetrics.radioClients.WithLabelValues(labels...).Set(float64(radioStat.NumClients))
		deviceMetrics.radioTransmitPowerDbm.WithLabelValues(labels...).Set(float64(radioStat.Power))
		deviceMetrics.radioReceiveBytesTotal.WithLabelValues(labels...).Set(float64(radioStat.RxBytes))
		deviceMetrics.radioReceivePacketsTotal.WithLabelValues(labels...).Set(float64(radioStat.RxPkts))
		deviceMetrics.radioTransmitBytesTotal.WithLabelValues(labels...).Set(float64(radioStat.TxBytes))
		deviceMetrics.radioTransmitPacketsTotal.WithLabelValues(labels...).Set(float64(radioStat.TxPkts))
	}
}
