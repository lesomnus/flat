package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lesomnus/flob"
	"github.com/lesomnus/xli/arg"
	"github.com/lesomnus/xli/flg"
)

type FlgReader = flg.Base[ReaderResolver, ReaderParser]
type ArgReader = arg.Mono[ReaderResolver, ReaderParser]

type ReaderParser string

func (ReaderParser) ToString(v ReaderResolver) string {
	return string(v)
}

func (ReaderParser) String() string {
	return "-|FILE"
}

func (ReaderParser) Parse(arg string) (ReaderResolver, error) {
	return ReaderResolver(arg), nil
}

type ReaderResolver string

func (d ReaderResolver) Resolve(r io.Reader) (io.ReadCloser, error) {
	if d == "-" {
		return io.NopCloser(r), nil
	}

	return os.Open(string(d))
}

type FlgDigest = flg.Base[DigestResolver, DigestParser]
type ArgDigest = arg.Mono[DigestResolver, DigestParser]

type DigestParser struct{}

func (DigestParser) ToString(v DigestResolver) string {
	return string(v)
}

func (DigestParser) String() string {
	return "-|FILE|HEX[32]"
}

func (DigestParser) Parse(arg string) (DigestResolver, error) {
	return DigestResolver(arg), nil
}

type DigestResolver string

func (d DigestResolver) Resolve(r io.Reader) (flob.Digest, error) {
	v := string(d)

	literal := false
	if strings.HasPrefix(v, ".") || strings.HasPrefix(v, "/") {
		f, err := os.Open(v)
		if err != nil {
			return "", fmt.Errorf("open file: %w", err)
		}
		defer f.Close()
		r = f
	} else if v == "-" {
		// Read from stdin
	} else {
		literal = true
	}

	if !literal {
		h := sha256.New()
		if _, err := io.Copy(h, r); err != nil {
			return "", fmt.Errorf("hash: %w", err)
		}
		v = hex.EncodeToString(h.Sum(nil))
	}

	return flob.Digest(v).Sanitize()
}
