package server

import (
	"context"

	"github.com/fmotalleb/timber/config"
)

// Context is the application context.
type Context interface {
	context.Context
	GetCfg() config.Config
}

type contextObj struct {
	context.Context
	cfg config.Config
}

// NewContext creates a new application context.
func NewContext(ctx context.Context, cfg config.Config) Context {
	return &contextObj{
		Context: ctx,
		cfg:     cfg,
	}
}

func (c *contextObj) GetCfg() config.Config {
	return c.cfg
}
