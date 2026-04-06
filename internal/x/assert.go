package x

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type X struct {
	t *testing.T
}

func New(t *testing.T) X {
	t.Helper()
	return X{t: t}
}

func (a X) Eq(got, want any) {
	a.t.Helper()
	if reflect.DeepEqual(got, want) {
		return
	}

	a.t.Fatalf("assert.Eq failed: got=%s want=%s", formatValue(got), formatValue(want))
}

func (a X) NotError(err error) {
	a.t.Helper()
	if err == nil {
		return
	}

	a.t.Fatalf("assert.NotError failed: err=%v", err)
}

func (a X) ErrorIs(err error, target error) {
	a.t.Helper()
	if errors.Is(err, target) {
		return
	}

	a.t.Fatalf("assert.ErrorIs failed: err=%v target=%v", err, target)
}

func formatValue(v any) string {
	return fmt.Sprintf("%#v", v)
}
