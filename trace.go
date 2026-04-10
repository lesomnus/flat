package flob

import (
	"context"
	"io"

	"github.com/lesomnus/otx"
)

var (
	_ Stores = TraceStores{}
	_ Store  = TraceStore{}
)

type TraceStores struct {
	Stores
}

func (t TraceStores) Use(id string) Store {
	return TraceStore{t.Stores.Use(id)}
}

type TraceStore struct {
	Store
}

func (t TraceStore) Add(ctx context.Context, m Meta, r io.Reader) (Meta, error) {
	ctx, span := otx.TraceStart(ctx, "add")
	defer span.End()

	return t.Store.Add(ctx, m, r)
}

func (t TraceStore) Erase(ctx context.Context, d Digest) error {
	ctx, span := otx.TraceStart(ctx, "erase")
	defer span.End()

	return t.Store.Erase(ctx, d)
}

func (t TraceStore) Get(ctx context.Context, d Digest) (Meta, error) {
	ctx, span := otx.TraceStart(ctx, "get")
	defer span.End()

	return t.Store.Get(ctx, d)
}

func (t TraceStore) Label(ctx context.Context, d Digest, labels Labels) error {
	ctx, span := otx.TraceStart(ctx, "label")
	defer span.End()

	return t.Store.Label(ctx, d, labels)
}

func (t TraceStore) Open(ctx context.Context, d Digest) (io.ReadSeekCloser, Meta, error) {
	ctx, span := otx.TraceStart(ctx, "open")
	defer span.End()

	return t.Store.Open(ctx, d)
}
