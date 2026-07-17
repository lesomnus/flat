package flob

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// signatureOf extracts the Signature= value from an Authorization header.
func signatureOf(t *testing.T, auth string) string {
	t.Helper()
	for _, part := range strings.Split(auth, ",") {
		part = strings.TrimSpace(part)
		if v, ok := strings.CutPrefix(part, "Signature="); ok {
			return v
		}
	}
	t.Fatalf("no Signature in Authorization header: %q", auth)
	return ""
}

// The credentials and expected signatures below are AWS's own published
// Signature Version 4 examples for S3 ("Authenticating Requests: Using the
// Authorization Header"), so they pin the signer to the reference algorithm.
const (
	exAccessKey = "AKIAIOSFODNN7EXAMPLE"
	exSecretKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
)

func exampleSigner() signer {
	return signer{
		creds:   Credentials{AccessKeyID: exAccessKey, SecretAccessKey: exSecretKey},
		region:  "us-east-1",
		service: "s3",
		now:     func() time.Time { return time.Date(2013, 5, 24, 0, 0, 0, 0, time.UTC) },
	}
}

func TestSigV4GetObjectExample(t *testing.T) {
	// GET /test.txt with a Range header.
	req := &http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Scheme: "https", Host: "examplebucket.s3.amazonaws.com", Opaque: "/test.txt"},
		Header: http.Header{},
	}
	req.Header.Set("Range", "bytes=0-9")

	exampleSigner().sign(req, emptyPayloadHash)

	const want = "f0e8bdb87c964420e857bd35b5d6ed310bd44f0170aba48dd91039c6036bdb41"
	if got := signatureOf(t, req.Header.Get("Authorization")); got != want {
		t.Fatalf("signature mismatch:\n got=%s\nwant=%s\nauth=%s", got, want, req.Header.Get("Authorization"))
	}
	if got := req.Header.Get("X-Amz-Date"); got != "20130524T000000Z" {
		t.Fatalf("x-amz-date=%q", got)
	}
}

func TestSigV4PutObjectExample(t *testing.T) {
	// PUT /test$file.text with a Date and a storage-class header, body
	// "Welcome to Amazon S3.".
	const bodyHash = "44ce7dd67c959e0d3524ffac1771dfbba87d2b6b4b4e99e42034a8b803f8b072"
	req := &http.Request{
		Method: http.MethodPut,
		URL:    &url.URL{Scheme: "https", Host: "examplebucket.s3.amazonaws.com", Opaque: "/test%24file.text"},
		Header: http.Header{},
	}
	req.Header.Set("Date", "Fri, 24 May 2013 00:00:00 GMT")
	req.Header.Set("X-Amz-Storage-Class", "REDUCED_REDUNDANCY")

	exampleSigner().sign(req, bodyHash)

	const want = "98ad721746da40c64f1a55b78f14c238d841ea1380cd77a1b5971af0ece108bd"
	if got := signatureOf(t, req.Header.Get("Authorization")); got != want {
		t.Fatalf("signature mismatch:\n got=%s\nwant=%s\nauth=%s", got, want, req.Header.Get("Authorization"))
	}
}

func TestSigV4SessionTokenHeaderSigned(t *testing.T) {
	s := exampleSigner()
	s.creds.SessionToken = "FQoGZXIvYXdzEXAMPLETOKEN"

	req := &http.Request{
		Method: http.MethodGet,
		URL:    &url.URL{Scheme: "https", Host: "examplebucket.s3.amazonaws.com", Opaque: "/test.txt"},
		Header: http.Header{},
	}
	s.sign(req, emptyPayloadHash)

	if req.Header.Get("X-Amz-Security-Token") == "" {
		t.Fatal("session token header not set")
	}
	if !strings.Contains(req.Header.Get("Authorization"), "x-amz-security-token") {
		t.Fatalf("session token not in signed headers: %s", req.Header.Get("Authorization"))
	}
}

func TestAwsURIEncode(t *testing.T) {
	cases := []struct {
		in          string
		encodeSlash bool
		want        string
	}{
		{"test.txt", false, "test.txt"},
		{"test$file.text", false, "test%24file.text"},
		{"a/b/c", false, "a/b/c"},
		{"a/b/c", true, "a%2Fb%2Fc"},
		{"key with space", true, "key%20with%20space"},
		{"tilde~_-.ok", true, "tilde~_-.ok"},
	}
	for _, c := range cases {
		if got := awsURIEncode(c.in, c.encodeSlash); got != c.want {
			t.Errorf("awsURIEncode(%q, %v)=%q want %q", c.in, c.encodeSlash, got, c.want)
		}
	}
}
