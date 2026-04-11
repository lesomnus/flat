package cmd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"

	"github.com/lesomnus/flob"
	"github.com/lesomnus/xli"
	"github.com/lesomnus/xli/arg"
	"github.com/lesomnus/z"
)

func NewCmdHash() *xli.Command {
	return &xli.Command{
		Name: "hash",

		Args: arg.Args{
			&ArgReader{Name: "FILE"},
		},

		Handler: xli.OnRun(func(ctx context.Context, cmd *xli.Command, next xli.Next) error {
			r, err := arg.MustGet[ReaderResolver](cmd, "FILE").Resolve(cmd)
			if err != nil {
				return z.Err(err, "resolve file")
			}
			defer r.Close()

			h := sha256.New()
			if _, err := io.Copy(h, r); err != nil {
				return z.Err(err, "hash")
			}

			cmd.Println(hex.EncodeToString(h.Sum(nil)))
			return next(ctx)
		}),
	}
}

func NewCmdAdd() *xli.Command {
	return &xli.Command{
		Name: "add",

		Args: arg.Args{
			&arg.String{Name: "STORE_ID"},
			&ArgReader{Name: "FILE"},
		},

		Handler: useClientStore(func(ctx context.Context, cmd *xli.Command, s flob.Stores) error {
			id := arg.MustGet[string](cmd, "STORE_ID")
			r, err := arg.MustGet[ReaderResolver](cmd, "FILE").Resolve(cmd)
			if err != nil {
				return z.Err(err, "resolve file")
			}
			defer r.Close()

			m, err := s.Use(id).Add(ctx, flob.Meta{}, r)
			if err != nil {
				if errors.Is(err, flob.ErrAlreadyExists) {
					cmd.Println(m.Digest)
				}
				return z.Err(err, "op")
			}

			cmd.Println(m.Digest)
			return nil
		}),
	}
}

func NewCmdGet() *xli.Command {
	return &xli.Command{
		Name: "get",

		Args: arg.Args{
			&arg.String{Name: "STORE_ID"},
			&ArgDigest{Name: "DIGEST"},
		},

		Handler: useClientStore(func(ctx context.Context, cmd *xli.Command, s flob.Stores) error {
			id := arg.MustGet[string](cmd, "STORE_ID")
			d, err := arg.MustGet[DigestResolver](cmd, "DIGEST").Resolve(cmd)
			if err != nil {
				return z.Err(err, "resolve digest")
			}

			m, err := s.Use(id).Get(ctx, d)
			if err != nil {
				return z.Err(err, "op")
			}

			cmd.Println("Digest:", m.Digest)
			cmd.Println("Size:", m.Size)
			if len(m.Labels) > 0 {
				cmd.Println("Labels:")
				for k, v := range m.Labels {
					cmd.Println("  ", k, "=", v)
				}
			}
			return nil
		}),
	}
}

func NewCmdRead() *xli.Command {
	return &xli.Command{
		Name: "read",

		Args: arg.Args{
			&arg.String{Name: "STORE_ID"},
			&ArgDigest{Name: "DIGEST"},
		},

		Handler: useClientStore(func(ctx context.Context, cmd *xli.Command, s flob.Stores) error {
			id := arg.MustGet[string](cmd, "STORE_ID")
			d, err := arg.MustGet[DigestResolver](cmd, "DIGEST").Resolve(cmd)
			if err != nil {
				return z.Err(err, "resolve digest")
			}

			f, _, err := s.Use(id).Open(ctx, d)
			if err != nil {
				return z.Err(err, "op")
			}
			defer f.Close()

			_, err = io.Copy(cmd, f)
			return err
		}),
	}
}

func NewCmdErase() *xli.Command {
	return &xli.Command{
		Name: "erase",
		Aliases: []string{
			"remove", "rm",
		},

		Args: arg.Args{
			&arg.String{Name: "STORE_ID"},
			&ArgDigest{Name: "DIGEST"},
		},

		Handler: useClientStore(func(ctx context.Context, cmd *xli.Command, s flob.Stores) error {
			id := arg.MustGet[string](cmd, "STORE_ID")
			d, err := arg.MustGet[DigestResolver](cmd, "DIGEST").Resolve(cmd)
			if err != nil {
				return z.Err(err, "resolve digest")
			}

			if err = s.Use(id).Erase(ctx, d); err != nil {
				return z.Err(err, "op")
			}
			return err
		}),
	}
}
