package cmd

import (
	"errors"

	"github.com/lesomnus/flob"
)

// Exit codes the CLI reports. Anything unclassified stays 1, so scripts that
// only test for failure are unaffected.
//
// The point is that a caller can tell "the store already has this" from "the
// upload failed" without parsing the message. A batch upload wants the former
// to be a skip: content is addressed by digest, so a blob that is already
// there is the desired end state, and re-running a publish is supposed to be
// cheap. Without a distinct code the only options are to swallow every failure
// or to grep stderr.
//
// 2 is deliberately left unused; it is conventionally a usage error.
const (
	// ExitFailure is any error the CLI does not classify.
	ExitFailure = 1
	// ExitAlreadyExists reports [flob.ErrAlreadyExists]: the blob is already
	// in the store, so nothing was written.
	ExitAlreadyExists = 3
	// ExitNotExist reports [flob.ErrNotExist]: the store answered, and the
	// blob is not there. Distinct from a transport or credential failure,
	// which stays [ExitFailure].
	ExitNotExist = 4
)

// ExitCode maps err to the process exit code that reports it.
func ExitCode(err error) int {
	switch {
	case err == nil:
		return 0
	case errors.Is(err, flob.ErrAlreadyExists):
		return ExitAlreadyExists
	case errors.Is(err, flob.ErrNotExist):
		return ExitNotExist
	default:
		return ExitFailure
	}
}
