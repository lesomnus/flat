package configs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/lesomnus/flob"
)

// echoBucket is a stand-in S3 endpoint: it answers HEAD with 404 so a Get returns
// ErrNotExist, which is enough to prove the config produced a working, signed
// client pointed at the right place.
func TestStoresConfigS3(t *testing.T) {
	var gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(404)
	}))
	t.Cleanup(srv.Close)

	doc := `
s3/main:
  endpoint: ` + srv.URL + `
  region: us-east-1
  bucket: my-bucket
  prefix: tenant-a
  path_style: true
  access_key_id: AKIAEXAMPLE
  secret_access_key: secretexample
`
	c := StoresConfig{}
	if err := yaml.Unmarshal([]byte(doc), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := c["s3/main"].(*StoresConfigS3); !ok {
		t.Fatalf("s3 config type = %T", c["s3/main"])
	}

	stores, err := c.build(context.Background(), "s3/main")
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	d := flob.Digest("sha256:0000000000000000000000000000000000000000000000000000000000000000")
	_, err = stores.Use("store-1").Get(context.Background(), d)
	if err == nil || !strings.Contains(err.Error(), "not exist") {
		t.Fatalf("get err = %v", err)
	}

	if !strings.Contains(gotPath, "/my-bucket/tenant-a/refs/") {
		t.Fatalf("request path = %q", gotPath)
	}
	if !strings.HasPrefix(gotAuth, "AWS4-HMAC-SHA256 ") {
		t.Fatalf("authorization = %q", gotAuth)
	}
}

func TestStoresConfigS3EnvFallback(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "env-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "env-secret")
	t.Setenv("AWS_REGION", "eu-west-1")

	doc := `
s3/env:
  bucket: b
  path_style: true
`
	c := StoresConfig{}
	if err := yaml.Unmarshal([]byte(doc), &c); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, err := c.build(context.Background(), "s3/env"); err != nil {
		t.Fatalf("build: %v", err)
	}
}
