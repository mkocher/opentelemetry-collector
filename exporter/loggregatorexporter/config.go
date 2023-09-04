// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package loggregatorexporter // import "go.opentelemetry.io/collector/exporter/loggregatorexporter"

import (
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

type Config struct {
	exporterhelper.TimeoutSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct.
	exporterhelper.QueueSettings   `mapstructure:"sending_queue"`
	exporterhelper.RetrySettings   `mapstructure:"retry_on_failure"`
	Endpoint                       string `mapstructure:"endpoint"`
	CaCert                         string `mapstructure:"ca_cert"`
	Cert                           string `mapstructure:"cert"`
	Key                            string `mapstructure:"key"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the exporter configuration is valid
func (cfg *Config) Validate() error {
	if err := cfg.QueueSettings.Validate(); err != nil {
		return fmt.Errorf("queue settings has invalid configuration: %w", err)
	}

	return nil
}
