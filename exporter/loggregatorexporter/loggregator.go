// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package loggregatorexporter // import "go.opentelemetry.io/collector/exporter/loggregatorexporter"

import (
	"context"
	"fmt"
	"log"

	"code.cloudfoundry.org/go-loggregator/v9"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

type baseExporter struct {
	// Input configuration.
	config *Config

	// gRPC clients and connection.
	client   *loggregator.IngressClient
	settings component.TelemetrySettings
}

// Crete new exporter and start it. The exporter will begin connecting but
// this function may return before the connection is established.
func newExporter(cfg component.Config, set exporter.CreateSettings) (*baseExporter, error) {
	oCfg := cfg.(*Config)

	return &baseExporter{config: oCfg, settings: set.TelemetrySettings}, nil
}

// start actually creates the gRPC connection. The client construction is deferred till this point as this
// is the only place we get hold of Extensions which are required to construct auth round tripper.
func (e *baseExporter) start(ctx context.Context, host component.Host) (err error) {
	tlsConfig, err := loggregator.NewIngressTLSConfig(
		e.config.CaCert,
		e.config.Cert,
		e.config.Key,
	)
	if err != nil {
		log.Fatal("Could not create TLS config", err)
	}

	client, err := loggregator.NewIngressClient(
		tlsConfig,
		loggregator.WithAddr(e.config.Endpoint),
	)

	if err != nil {
		log.Fatal("Could not create client", err)
	}

	e.client = client
	return
}

func (e *baseExporter) shutdown(context.Context) error {
	if e.client != nil {
		return e.client.CloseSend()
	}
	return nil
}

func (e *baseExporter) pushMetrics(ctx context.Context, md pmetric.Metrics) error {
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				switch t := metric.Type(); t {
				case pmetric.MetricTypeGauge:
					dp := metric.Gauge().DataPoints().At(0)
					if dp.ValueType() == pmetric.NumberDataPointValueTypeDouble {
						fmt.Printf("Sending a gauge to Loggregator: %s => %f", metric.Name(), metric.Gauge().DataPoints().At(0).DoubleValue())
						e.client.EmitGauge(loggregator.WithGaugeValue(metric.Name(), metric.Gauge().DataPoints().At(0).DoubleValue(), metric.Unit()), loggregator.WithGaugeSourceInfo("OTEL1", "OTEL2"))
					} else {
						fmt.Printf("Sending a gauge to Loggregator: %s => %f", metric.Name(), float64(metric.Gauge().DataPoints().At(0).IntValue()))
						e.client.EmitGauge(loggregator.WithGaugeValue(metric.Name(), float64(metric.Gauge().DataPoints().At(0).IntValue()), metric.Unit()), loggregator.WithGaugeSourceInfo("OTEL1", "OTEL2"))
					}
				case pmetric.MetricTypeSum:
					dp := metric.Sum().DataPoints().At(0)
					if dp.ValueType() == pmetric.NumberDataPointValueTypeDouble {
						fmt.Printf("Sendng a sum to Loggregator: %s => %d", metric.Name(), uint64(dp.DoubleValue()))
						e.client.EmitCounter(metric.Name(), loggregator.WithDelta(uint64(dp.DoubleValue())), loggregator.WithCounterSourceInfo("OTEL1", "OTEL2"))
					} else {
						fmt.Printf("Sendng a sum to Loggregator: %s => %d", metric.Name(), uint64(dp.IntValue()))
						e.client.EmitCounter(metric.Name(), loggregator.WithDelta(uint64(dp.IntValue())), loggregator.WithCounterSourceInfo("OTEL1", "OTEL2"))
					}
				}
			}
		}
	}
	return nil
}

func (e *baseExporter) pushLogs(ctx context.Context, ld plog.Logs) error {
	return nil
}
