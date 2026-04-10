package configs

import (
	"github.com/lesomnus/z"
)

type ClientConfig struct {
	Use string
}

func (c *ClientConfig) Evaluate() error {
	z.FallbackP(&c.Use, "http/local")

	return nil
}
