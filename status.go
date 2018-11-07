package main

import (
	_ "net/http/pprof"

	vault_api "github.com/hashicorp/vault/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

const (
	namespace = "vault"
)

var (
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last query of Vault successful.",
		nil, nil,
	)
	initialized = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "initialized"),
		"Is the Vault initialised (according to this node).",
		nil, nil,
	)
	sealed = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "sealed"),
		"Is the Vault node sealed.",
		nil, nil,
	)
	standby = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "standby"),
		"Is this Vault node in standby.",
		nil, nil,
	)
	info = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "info"),
		"Version of this Vault node.",
		[]string{"version", "cluster_name", "cluster_id"}, nil,
	)
)

// StatusExporter collects Vault health from the given server and exports them using
// the Prometheus metrics package.
type StatusExporter struct {
	client *vault_api.Client
}

// NewStatusExporter returns an initialized StatusExporter.
func NewStatusExporter() (*StatusExporter, error) {
	vaultConfig := vault_api.DefaultConfig()

	tlsconfig := &vault_api.TLSConfig{
		Insecure: true,
	}
	err := vaultConfig.ConfigureTLS(tlsconfig)
	if err != nil {
		return nil, err
	}

	client, err := vault_api.NewClient(vaultConfig)
	if err != nil {
		return nil, err
	}

	return &StatusExporter{
		client: client,
	}, nil
}

// Describe describes all the metrics ever exported by the Vault exporter. It
// implements prometheus.Collector.
func (e *StatusExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- initialized
	ch <- sealed
	ch <- standby
	ch <- info
}

func bool2float(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// Collect fetches the stats from configured Vault and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *StatusExporter) Collect(ch chan<- prometheus.Metric) {
	health, err := e.client.Sys().Health()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0,
		)
		log.Errorf("Failed to collect health from Vault server: %v", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, 1,
	)
	ch <- prometheus.MustNewConstMetric(
		initialized, prometheus.GaugeValue, bool2float(health.Initialized),
	)
	ch <- prometheus.MustNewConstMetric(
		sealed, prometheus.GaugeValue, bool2float(health.Sealed),
	)
	ch <- prometheus.MustNewConstMetric(
		standby, prometheus.GaugeValue, bool2float(health.Standby),
	)
	ch <- prometheus.MustNewConstMetric(
		info, prometheus.GaugeValue, 1, health.Version, health.ClusterName, health.ClusterID,
	)
}
