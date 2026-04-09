package configs

import (
	"errors"

	"github.com/lesomnus/xddr"
	"github.com/lesomnus/z"
)

type ServerConfig struct {
	Use  string
	Addr xddr.HTTPLocal
}

func (c *ServerConfig) Evaluate() error {
	z.FallbackP(&c.Use, "mem")
	z.FallbackP(&c.Addr, "0.0.0.0:8080")

	return errors.Join(
		z.ErrIf(z.Take(c.Addr.Sanitize()).To(&c.Addr), ".addr"),
	)
}
