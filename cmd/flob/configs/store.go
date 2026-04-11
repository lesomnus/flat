package configs

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lesomnus/flob"
	"github.com/lesomnus/otx/otxhttp"
	"github.com/lesomnus/z"
)

type StoresConfig map[string]any

func (c StoresConfig) Use(ctx context.Context, name string) (flob.Stores, error) {
	s, err := c.build(ctx, name)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (c StoresConfig) build(ctx context.Context, k string) (s flob.Stores, err error) {
	defer func() {
		if s != nil {
			s = StoresTrace{Stores: s}
		}
	}()

	if k == "mem" {
		return flob.NewMemStores(), nil
	}

	v, ok := c[k]
	if !ok {
		return nil, fmt.Errorf("unknown store: %q", k)
	}

	switch c_ := v.(type) {
	case *StoresConfigMem:
		return flob.NewMemStores(), nil

	case *StoresConfigOs:
		return flob.NewOsStores(c_.Path), nil

	case *StoresConfigHttp:
		client := *http.DefaultClient
		client.Transport = otxhttp.NewTransport(client.Transport)

		return flob.HttpStores{
			Client: &client,
			Target: c_.Target,
		}, nil

	case *StoresConfigCache:
		primary, err := c.build(ctx, c_.Primary)
		if err != nil {
			return nil, z.Err(err, "build primary store: %q", c_.Primary)
		}

		origin, err := c.build(ctx, c_.Origin)
		if err != nil {
			return nil, z.Err(err, "build origin store: %q", c_.Origin)
		}
		return flob.CacheStores{
			Primary: primary,
			Origin:  origin,
		}, nil

	default:
		return nil, fmt.Errorf("invalid store config: %T", v)
	}
}

func (c StoresConfig) UnmarshalYAML(f func(v any) error) error {
	type C map[string]any

	c_ := C{}
	if err := f(&c_); err != nil {
		return err
	}
	cm := configmap{
		"mem":   StoresConfigMem{},
		"os":    StoresConfigOs{},
		"cache": StoresConfigCache{},
		"http":  StoresConfigHttp{},
	}
	for k, v := range c_ {
		parts := strings.SplitN(k, "/", 2)
		if len(parts) == 0 {
			return fmt.Errorf("invalid key: %q", k)
		}

		kind := parts[0]
		var err error
		c[k], err = cm.unmarshal(kind, v)
		if err != nil {
			return z.Err(err, "unmarshal: %q", k)
		}
	}
	return nil
}

type StoresConfigMem struct{}

type StoresConfigOs struct {
	Path string
}

type StoresConfigHttp struct {
	Target string
}

type StoresConfigCache struct {
	Primary string
	Origin  string
}
