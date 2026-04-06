package flat_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/lesomnus/flat"
	"github.com/lesomnus/flat/internal/x"
)

type newStoreFn func(t *testing.T) flat.Store

func testStore(t *testing.T, new_store newStoreFn) {
	t.Helper()

	data := []byte("Royale with Cheese")
	new_reader := func() io.Reader {
		return bytes.NewReader(data)
	}

	t.Run("add then get", func(t *testing.T) {
		x := x.New(t)
		s := new_store(t)

		labels := flat.Labels{"Content-Type": {"text/plain"}}
		added, err := s.Add(t.Context(), flat.Meta{Labels: labels}, new_reader())
		x.NotError(err)

		got, err := s.Get(t.Context(), added.Digest)
		x.NotError(err)
		x.Eq(got.Digest, added.Digest)
		x.Eq(got.Size, int64(len(data)))
	})
	t.Run("add duplicate returns ErrAlreadyExists", func(t *testing.T) {
		x := x.New(t)
		s := new_store(t)

		_, err := s.Add(t.Context(), flat.Meta{}, new_reader())
		x.NotError(err)

		_, err = s.Add(t.Context(), flat.Meta{}, new_reader())
		x.ErrorIs(err, flat.ErrAlreadyExists)
	})
	t.Run("get missing returns ErrNotExist", func(t *testing.T) {
		x := x.New(t)
		s := new_store(t)

		_, err := s.Get(t.Context(), flat.Digest("0000000000000000000000000000000000000000000000000000000000000000"))
		x.ErrorIs(err, flat.ErrNotExist)
	})
	t.Run("open returns same content after add", func(t *testing.T) {
		x := x.New(t)
		s := new_store(t)

		m, err := s.Add(t.Context(), flat.Meta{}, new_reader())
		x.NotError(err)

		r, _, err := s.Open(t.Context(), m.Digest)
		x.NotError(err)
		defer r.Close()

		got, err := io.ReadAll(r)
		x.NotError(err)
		x.Eq(got, data)
	})
}
