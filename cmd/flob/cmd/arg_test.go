package cmd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestDigestResolver(t *testing.T) {
	const content = "Royale with Cheese"
	want := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(content)))
	bareHex := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))

	t.Run("from content on stdin", func(t *testing.T) {
		got, err := DigestResolver("-").Resolve(bytes.NewReader([]byte(content)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != want {
			t.Fatalf("got=%q want=%q", got, want)
		}
	})
	t.Run("literal bare hex is accepted as sha256", func(t *testing.T) {
		got, err := DigestResolver(bareHex).Resolve(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != want {
			t.Fatalf("got=%q want=%q", got, want)
		}
	})
	t.Run("literal prefixed digest is accepted", func(t *testing.T) {
		got, err := DigestResolver(want).Resolve(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != want {
			t.Fatalf("got=%q want=%q", got, want)
		}
	})
}
