package server

import (
	"context"

	"github.com/fmotalleb/timber/config"
)

type Context interface {
	context.Context
	GetCfg() config.Config
}

type contextObj struct {
	context.Context
	cfg config.Config
}

func NewContext(ctx context.Context, cfg config.Config) Context {
	return &contextObj{
		Context: ctx,
		cfg:     cfg,
	}
}

func (c *contextObj) GetCfg() config.Config {
	return c.cfg
}
