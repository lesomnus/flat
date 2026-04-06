# flat

A content-addressable storage (CAS) backed purely by the filesystem — no database required.
Blobs are identified by their SHA-256 digest and organized under isolated 1-depth namespaced stores.

## Usage

```go
import "github.com/lesomnus/flat"

func main() {
	stores := flat.NewOsStores("/path/to/storage")

	// Store "foo" and "bar" are independent namespaces but data are deduplicated 
	// and shared across them if the same content is added.
	store_foo := stores.Use("foo")
	store_bar := stores.Use("bar")

	ctx := context.Background()
	meta, _ := store_foo.Add(ctx, []byte("hello world"))

	r, _, _ := store_foo.Open(ctx, meta.Digest) 
	io.ReadAll(r) // "hello world"

	_, _, err := store_bar.Open(ctx, meta.Digest)
	err // flat.ErrNotExist

	// same content, same digest, no duplicate storage.
	store_bar.Add(ctx, []byte("hello world")) 
}

```
