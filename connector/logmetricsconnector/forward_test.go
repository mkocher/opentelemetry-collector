// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0
package logmetricsconnector

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/connector/connectortest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestForward(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig()
	assert.Equal(t, &struct{}{}, cfg)

	ctx := context.Background()
	set := connectortest.NewNopCreateSettings()
	host := componenttest.NewNopHost()

	metricsSink := new(consumertest.MetricsSink)
	logsToMetrics, err := f.CreateLogsToMetrics(ctx, set, cfg, metricsSink)
	assert.NoError(t, err)
	assert.NotNil(t, logsToMetrics)

	assert.NoError(t, logsToMetrics.Start(ctx, host))

	assert.NoError(t, logsToMetrics.ConsumeLogs(ctx, plog.NewLogs()))
	assert.NoError(t, logsToMetrics.ConsumeLogs(ctx, plog.NewLogs()))

	assert.NoError(t, logsToMetrics.Shutdown(ctx))

	assert.Equal(t, 0, len(metricsSink.AllMetrics()))
}

func TestJson(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig()
	assert.Equal(t, &struct{}{}, cfg)

	ctx := context.Background()
	set := connectortest.NewNopCreateSettings()
	host := componenttest.NewNopHost()

	metricsSink := new(consumertest.MetricsSink)
	logsToMetrics, err := f.CreateLogsToMetrics(ctx, set, cfg, metricsSink)
	assert.NoError(t, err)
	assert.NotNil(t, logsToMetrics)

	assert.NoError(t, logsToMetrics.Start(ctx, host))

	logs := plog.NewLogs()
	_ = logs.ResourceLogs().AppendEmpty().Resource()
	log := logs.ResourceLogs().At(0).ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	log.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	log.SetDroppedAttributesCount(1)
	log.SetSeverityNumber(plog.SeverityNumberInfo)
	log.SetSeverityText("Info")

	log.Body().SetStr("{\"type\": \"counter\",\"name\": \"some-counter\", \"delta\": 7, \"tags\": { \"tag1\": \"tag value\" }}")
	assert.NoError(t, logsToMetrics.ConsumeLogs(ctx, logs))

	log.Body().SetStr("{\"type\": \"gauge\",\"name\": \"some-gauge\", \"value\": 11, \"tags\": { \"tag1\": \"tag value\" }}")
	assert.NoError(t, logsToMetrics.ConsumeLogs(ctx, logs))

	assert.NoError(t, logsToMetrics.Shutdown(ctx))

	assert.Equal(t, 2, len(metricsSink.AllMetrics()))

	metric := metricsSink.AllMetrics()[0].ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	assert.Equal(t, "some-counter", metric.Name())
	assert.Equal(t, int64(7), metric.Sum().DataPoints().At(0).IntValue())
	assert.Equal(t, pmetric.AggregationTemporalityDelta, metric.Sum().AggregationTemporality())

	metric = metricsSink.AllMetrics()[1].ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0)
	assert.Equal(t, "some-gauge", metric.Name())
	assert.Equal(t, int64(11), metric.Gauge().DataPoints().At(0).IntValue())
	// TODO: tags
}
