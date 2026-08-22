package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/lesomnus/flob"
)

func TestExitCode(t *testing.T) {
	tcs := []struct {
		desc string
		err  error
		want int
	}{
		{"nil is success", nil, 0},
		{"already exists", flob.ErrAlreadyExists, ExitAlreadyExists},
		{"not exist", flob.ErrNotExist, ExitNotExist},
		// The handlers wrap twice before main sees the error.
		{"wrapped already exists", fmt.Errorf("run command: %w", fmt.Errorf("op: %w", flob.ErrAlreadyExists)), ExitAlreadyExists},
		{"wrapped not exist", fmt.Errorf("run command: %w", fmt.Errorf("op: %w", flob.ErrNotExist)), ExitNotExist},
		{"unclassified", errors.New("connection reset by peer"), ExitFailure},
		{"digest mismatch is not a skip", flob.ErrDigestMismatch, ExitFailure},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			if got := ExitCode(tc.err); got != tc.want {
				t.Fatalf("got=%d want=%d", got, tc.want)
			}
		})
	}
}
