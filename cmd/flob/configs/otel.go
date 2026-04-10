package configs

import (
	"context"
	"errors"

	"github.com/lesomnus/flob/cmd/flob/version"
	"github.com/lesomnus/mkot"
	"github.com/lesomnus/otx"
	"github.com/lesomnus/z"
	"go.opentelemetry.io/otel/attribute"
	nooplog "go.opentelemetry.io/otel/log/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	_ "github.com/lesomnus/mkot/otlp"
	"github.com/lesomnus/mkot/pretty"
	_ "github.com/lesomnus/mkot/pretty"
)

type OtelConfig struct {
	mkot.Config `yaml:",inline"`
}

func (c *OtelConfig) Build(ctx context.Context) (context.Context, *otx.Otx, error) {
	otc := mkot.NewConfig()
	if c != nil {
		otc = &c.Config
	}

	const ServiceResourceId mkot.Id = "resource/flob"
	if otc.Processors == nil {
		otc.Processors = map[mkot.Id]mkot.ProcessorConfig{}
	}
	if otc.Exporters == nil {
		otc.Exporters = map[mkot.Id]mkot.ExporterConfig{}
	}
	if otc.Processors == nil {
		otc.Processors = map[mkot.Id]mkot.ProcessorConfig{}
	}
	if otc.Providers == nil {
		otc.Providers = map[mkot.Id]*mkot.ProviderConfig{}
	}
	otc.Processors[ServiceResourceId] = &mkot.Resource{
		Attributes: []mkot.Attr{
			{Key: "service.name", Value: attribute.StringValue("flob")},
			{Key: "service.version", Value: attribute.StringValue(version.Get().Version)},
		},
	}
	if len(otc.Providers) == 0 {
		id := mkot.Id("pretty")
		if _, ok := otc.Exporters[id]; !ok {
			otc.Exporters[id] = pretty.ExporterConfig{}
		}
		otc.Providers["logger"] = &mkot.ProviderConfig{
			Exporters: []mkot.Id{id},
		}
	}

	for k := range otc.Providers {
		otc.Providers[k].Processors = append(otc.Providers[k].Processors, ServiceResourceId)
	}

	resolver := mkot.Make(ctx, otc)

	tracker_provider, err := resolver.Tracer(ctx, "")
	if err != nil {
		if !errors.Is(err, mkot.ErrNotExist) {
			return nil, nil, z.Err(err, "resolve tracer provider")
		}
		tracker_provider = sdktrace.NewTracerProvider()
	}

	// meter_provider, err := resolver.Meter(ctx, "")
	// if err != nil {
	// 	if !errors.Is(err, mkot.ErrNotExist) {
	// 		 nil,return nil, z.Err(err, "resolve meter provider")
	// 	}
	// 	meter_provider = noopmetric.NewMeterProvider()
	// }

	logger_provider, err := resolver.Logger(ctx, "")
	if err != nil {
		if !errors.Is(err, mkot.ErrNotExist) {
			return nil, nil, z.Err(err, "resolve logger provider")
		}
		logger_provider = nooplog.NewLoggerProvider()
	}
	v := otx.New(
		otx.WithController(resolver),
		otx.WithTracerProvider(tracker_provider),
		// otx.WithMeterProvider(meter_provider),
		otx.WithLoggerProvider(logger_provider),
	)
	return otx.Into(ctx, v), v, nil
}
