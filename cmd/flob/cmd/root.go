package cmd

import (
	"context"

	"github.com/lesomnus/flob"
	"github.com/lesomnus/xli"
	"github.com/lesomnus/z"
)

func NewCmdRoot() *xli.Command {
	return &xli.Command{
		Name: "flob",

		Commands: []*xli.Command{
			NewCmdConf(),
			NewCmdServe(),
			NewCmdAdd(),
			NewCmdGet(),
			NewCmdRead(),
		},

		Handler: xli.Chain(
			xli.RequireSubcommand(),
			configHandler(),
		),
	}
}

func useClientStore(f func(ctx context.Context, cmd *xli.Command, s flob.Stores) error) xli.Handler {
	return xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
		c := use_config.Must(ctx)

		s, err := c.Stores.Use(ctx, c.Client.Use)
		if err != nil {
			return z.Err(err, "build stores")
		}
		if err := f(ctx, cmd, s); err != nil {
			return z.Err(err, "run command")
		}
		return next(ctx)
	})
}
