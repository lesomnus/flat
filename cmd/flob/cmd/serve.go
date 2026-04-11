package cmd

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/lesomnus/flob"
	"github.com/lesomnus/flob/cmd/flob/configs"
	"github.com/lesomnus/otx"
	"github.com/lesomnus/otx/log"
	"github.com/lesomnus/otx/otxhttp"
	"github.com/lesomnus/xli"
	"github.com/lesomnus/z"
)

func NewCmdServe() *xli.Command {
	return &xli.Command{
		Name: "serve",
		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			c := use_config.Must(ctx)
			s, err := c.Stores.Use(ctx, c.Server.Use)
			if err != nil {
				return z.Err(err, "use stores: %s", c.Server.Use)
			}

			s, err = configs.NewStoresMeter(ctx, s)
			if err != nil {
				return z.Err(err, "new stores meter")
			}

			l := log.From(ctx)

			h := flob.HttpHandler{Stores: s}
			mux := http.NewServeMux()
			mux.Handle("/", otxhttp.NewHandler(otx.From(ctx), otxhttp.BoundaryLogger()(h), "/"))

			l.Info("serve", slog.String("addr", ":8080"))
			if err := http.ListenAndServe(":8080", mux); err != nil {
				return z.Err(err, "start http server")
			}

			return nil
		}),
	}
}
