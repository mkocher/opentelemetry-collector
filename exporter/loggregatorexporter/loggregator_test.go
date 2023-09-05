// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package loggregatorexporter // import "go.opentelemetry.io/collector/exporter/loggregatorexporter"

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestSendMetrics(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	// Disable queuing to ensure that we execute the request when calling ConsumeMetrics
	// otherwise we will not see any errors.
	cfg.QueueSettings.Enabled = false
	cfg.CaCert = "/home/pivotal/workspace/opentelemetry-collector/exporter/loggregatorexporter/testdata/test_cert.pem"
	cfg.Cert = "/home/pivotal/workspace/opentelemetry-collector/exporter/loggregatorexporter/testdata/test_cert.pem"
	cfg.Key = "/home/pivotal/workspace/opentelemetry-collector/exporter/loggregatorexporter/testdata/test_key.pem"
	cfg.Endpoint = "1.2.3.4:1234"

	set := exportertest.NewNopCreateSettings()
	set.BuildInfo.Description = "Collector"
	set.BuildInfo.Version = "1.2.3test"
	exp, err := factory.CreateMetricsExporter(context.Background(), set, cfg)
	require.NoError(t, err)
	require.NotNil(t, exp)
	defer func() {
		assert.NoError(t, exp.Shutdown(context.Background()))
	}()

	host := componenttest.NewNopHost()

	assert.NoError(t, exp.Start(context.Background(), host))

	// Send empty metric.
	md := pmetric.NewMetrics()
	metric := md.ResourceMetrics().AppendEmpty().ScopeMetrics().AppendEmpty().Metrics().AppendEmpty()
	metric.SetName("metric-name")
	metric.SetEmptySum()
	metric.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityDelta)
	metric.Sum().DataPoints().AppendEmpty().SetIntValue(42)
	assert.NoError(t, exp.ConsumeMetrics(context.Background(), md))

}
