// Copyright 2013 The KubeVault Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	_ "net/http/pprof"

	vault_api "github.com/hashicorp/vault/api"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
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

var (
	vaultCACert = kingpin.Flag("vault.tls-cacert",
		"The path to a PEM-encoded CA cert file to use to verify the Vault server SSL certificate.").String()
	vaultClientCert = kingpin.Flag("vault.tls-client-cert",
		"The path to the certificate for Vault communication.").String()
	vaultClientKey = kingpin.Flag("vault.tls-client-key",
		"The path to the private key for Vault communication.").String()
	insecureSkipVerify = kingpin.Flag("vault.insecure-skip-verify",
		"Set SSL to ignore certificate validation.").
		Default("false").Bool()
)

// StatusExporter collects Vault health from the given server and exports them using
// the Prometheus metrics package.
type StatusExporter struct {
	client *vault_api.Client
}

// NewStatusExporter returns an initialized StatusExporter.
func NewStatusExporter() (*StatusExporter, error) {
	vaultConfig := vault_api.DefaultConfig()

	if *insecureSkipVerify {
		tlsconfig := &vault_api.TLSConfig{
			Insecure: true,
		}
		err := vaultConfig.ConfigureTLS(tlsconfig)
		if err != nil {
			return nil, err
		}
	}

	if *vaultCACert != "" || *vaultClientCert != "" || *vaultClientKey != "" {

		tlsconfig := &vault_api.TLSConfig{
			CACert:     *vaultCACert,
			ClientCert: *vaultClientCert,
			ClientKey:  *vaultClientKey,
			Insecure:   *insecureSkipVerify,
		}
		err := vaultConfig.ConfigureTLS(tlsconfig)
		if err != nil {
			return nil, err
		}
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
