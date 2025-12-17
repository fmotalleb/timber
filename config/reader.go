package config

import (
	"context"
	"fmt"

	"github.com/fmotalleb/go-tools/config"
	"github.com/fmotalleb/go-tools/decoder"
	"github.com/fmotalleb/go-tools/defaulter"
)

// Parse reads and merges configuration from the given path and decodes it into the dst struct.
func Parse(ctx context.Context, dst *Config, path string) error {
	cfg, err := config.ReadAndMergeConfig(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to read and merge configs: %w", err)
	}
	decoder, err := decoder.Build(dst)
	if err != nil {
		return fmt.Errorf("create decoder: %w", err)
	}

	if err := decoder.Decode(cfg); err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	defaulter.ApplyDefaults(dst, dst)
	return nil
}
