// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package loggregatorexporter // import "go.opentelemetry.io/collector/exporter/loggregatorexporter"

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/exporter/exportertest"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg, "failed to create default config")
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
	ocfg, ok := factory.CreateDefaultConfig().(*Config)
	assert.True(t, ok)
	assert.Equal(t, ocfg.RetrySettings, exporterhelper.NewDefaultRetrySettings())
	assert.Equal(t, ocfg.QueueSettings, exporterhelper.NewDefaultQueueSettings())
	assert.Equal(t, ocfg.TimeoutSettings, exporterhelper.NewDefaultTimeoutSettings())
}

func TestCreateMetricsExporter(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	// cfg.GRPCClientSettings.Endpoint = testutil.GetAvailableLocalAddress(t)

	set := exportertest.NewNopCreateSettings()
	oexp, err := factory.CreateMetricsExporter(context.Background(), set, cfg)
	require.Nil(t, err)
	require.NotNil(t, oexp)
}
