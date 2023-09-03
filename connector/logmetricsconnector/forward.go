// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package logmetricsconnector // import "go.opentelemetry.io/collector/connector/forwardconnector"

import (
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/consumer"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

const (
	typeStr = "logmetrics"
)

// NewFactory returns a connector.Factory.
func NewFactory() connector.Factory {
	return connector.NewFactory(
		typeStr,
		createDefaultConfig,
		connector.WithLogsToMetrics(createLogsToMetrics, component.StabilityLevelBeta),
	)
}

// createDefaultConfig creates the default configuration.
func createDefaultConfig() component.Config {
	return &struct{}{}
}

func createLogsToMetrics(
	_ context.Context,
	_ connector.CreateSettings,
	_ component.Config,
	nextConsumer consumer.Metrics,
) (connector.Logs, error) {
	return &logmetrics{Metrics: nextConsumer}, nil
}

// forward is used to pass signals directly from one pipeline to another.
// This is useful when there is a need to replicate data and process it in more
// than one way. It can also be used to join pipelines together.
type logmetrics struct {
	consumer.Metrics
	component.StartFunc
	component.ShutdownFunc
}

func (c *logmetrics) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

type jsonLogmetric struct {
	Kind  string            `json:"type"`
	Name  string            `json:"name"`
	Value int64             `json:"value"`
	Delta int64             `json:"delta"`
	Tags  map[string]string `json:"tags"`
}

func (c *logmetrics) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	logmetric := jsonLogmetric{}

	for rl := 0; rl < logs.ResourceLogs().Len(); rl++ {
		resourcelog := logs.ResourceLogs().At(rl)
		for sl := 0; sl < resourcelog.ScopeLogs().Len(); sl++ {
			scopelog := resourcelog.ScopeLogs().At(sl)
			for lr := 0; lr < scopelog.LogRecords().Len(); lr++ {
				logrecord := scopelog.LogRecords().At(lr)
				err := json.Unmarshal([]byte(logrecord.Body().AsString()), &logmetric)
				if err == nil {
					m := pmetric.NewMetrics()
					rm := m.ResourceMetrics().AppendEmpty()
					sm := rm.ScopeMetrics().AppendEmpty()
					// sm.Scope().SetName("logmetricsconnector")
					metric := sm.Metrics().AppendEmpty()
					metric.SetName(logmetric.Name)

					switch k := logmetric.Kind; k {
					case "counter":
						metric.SetEmptySum()
						metric.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
						dp := metric.Sum().DataPoints().AppendEmpty()
						dp.SetIntValue(logmetric.Delta)
						_ = c.Metrics.ConsumeMetrics(ctx, m)
					case "gauge":
						metric.SetEmptyGauge()
						dp := metric.Gauge().DataPoints().AppendEmpty()
						dp.SetIntValue(logmetric.Value)
						_ = c.Metrics.ConsumeMetrics(ctx, m)
					case "event":
						fmt.Println("events are not supported")
					}
				}
			}
		}
	}

	return nil
}
