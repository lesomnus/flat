package flat

// /
// ├─ stage/
// │  └─ (random)/
// │      ├- blob
// │      └- labels
// ├─ locks/
// │  └─ xxxxx...
// ├─ blobs/
// │  └─ xx/
// │     └─ xx/
// │        └─ xxxx...
// └─ index/
//    └─ (id)/
//       └─ xx/
//          └─ xx/
//             └─ xxxx.../
//                ├- blob
//                └- labels

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

var (
	_ Stores = OsStores{}
	_ Store  = OsStore{}
)

type OsStores struct {
	root string
}

func NewOsStores(root string) OsStores {
	return OsStores{root}
}

func (i OsStores) Root() string {
	return i.root
}

func (i OsStores) Use(id string) Store {
	return OsStore{
		root: i.root,
		repo: filepath.Join(i.root, "index", id),
	}
}

type OsStore struct {
	root string
	repo string
}

func (s OsStore) Add(ctx context.Context, m Meta, r io.Reader) (Meta, error) {
	if m.Digest != "" {
		// Digest is provided, so check if the blob already exists.
		d, err := m.Digest.Sanitize()
		if err != nil {
			return m, err
		}
		m.Digest = d

		pb := s.pathToRepo(m.Digest, "blob")
		if err := s.checkDup(pb); err != nil {
			return m, err
		}

		// Blob with the given digest does not exist, so proceed to add it.
	}

	// Write to a local temp file first, not to the store root since the store root
	// may be on a different device and we need to compute the digest while writing.
	tf, err := os.CreateTemp("", "flat-*")
	if err != nil {
		return m, fmt.Errorf("create temp: %w", err)
	}

	tp := tf.Name()
	defer os.Remove(tp)
	defer tf.Close()

	h := Hash()
	n, err := io.Copy(io.MultiWriter(tf, h), r)
	if err != nil {
		return m, fmt.Errorf("write temp blob: %w", err)
	}
	if _, err := tf.Seek(0, io.SeekStart); err != nil {
		return m, fmt.Errorf("seek temp blob: %w", err)
	}

	m.Size = n

	d := Digest(fmt.Sprintf("%x", h.Sum(nil)))
	pb := s.pathToRepo(d, "blob")
	if m.Digest == "" {
		m.Digest = d
		// Now we have the digest, so check if the blob already exists to avoid
		// unnecessary work.
		if err := s.checkDup(pb); err != nil {
			return m, err
		}
	} else if m.Digest != d {
		return m, ErrDigestMismatch
	}

	// We decided to add the blob, so acquire the lock to prevent concurrent Add
	// or Erase with the same digest.
	unlock, err := s.lock(ctx, d)
	if err != nil {
		return m, err
	}
	defer unlock()

	// Maybe another process added the blob while we were waiting for the lock, so
	// check again.
	if err := s.checkDup(pb); err != nil {
		return m, err
	}

	// We are the only one adding the blob with the given digest, so stage the blob.
	ps := filepath.Join(s.root, "stage")
	if err := os.MkdirAll(ps, 0o755); err != nil {
		return m, fmt.Errorf("mkdir stage: %w", err)
	}

	ps, err = os.MkdirTemp(ps, "")
	if err != nil {
		return m, fmt.Errorf("mkdir temp at stage: %w", err)
	}
	defer os.RemoveAll(ps)

	// Write labels first.
	lf, err := os.Create(filepath.Join(ps, "labels"))
	if err != nil {
		return m, fmt.Errorf("create labels: %w", err)
	}
	if err := writeLabels(lf, m.Labels); err != nil {
		lf.Close()
		return m, fmt.Errorf("write labels: %w", err)
	}
	if err := lf.Close(); err != nil {
		return m, fmt.Errorf("close labels: %w", err)
	}

	pd := s.pathToBlob(d)
	if _, err := os.Stat(pd); err == nil {
		// There is already a blob with the same digest, so we can just make a hard link
		// to the destination path.
		if err := os.Link(pd, filepath.Join(ps, "blob")); err != nil {
			return m, fmt.Errorf("link blob: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return m, fmt.Errorf("stat blob: %w", err)
	} else {
		// No blob in the global namespace, so copy the temp file to staging area.
		psb := filepath.Join(ps, "blob")
		bf, err := os.Create(psb)
		if err != nil {
			return m, fmt.Errorf("create blob: %w", err)
		}
		defer bf.Close()

		if _, err := io.Copy(bf, tf); err != nil {
			return m, fmt.Errorf("copy blob: %w", err)
		}

		// Make a hard link to the global blob path.
		if err := os.MkdirAll(filepath.Dir(pd), 0o755); err != nil {
			return m, fmt.Errorf("mkdir blob dir: %w", err)
		}
		if err := os.Link(psb, pd); err != nil {
			return m, fmt.Errorf("link blob to global path: %w", err)
		}
	}

	// Now the blob is staged, so move it to the destination path atomically.
	if err := os.MkdirAll(filepath.Dir(pb), 0o755); err != nil {
		return m, fmt.Errorf("mkdir repo: %w", err)
	}
	if err := os.Rename(ps, pb); err != nil {
		return m, fmt.Errorf("move from stage to repo: %w", err)
	}

	return m, nil
}

func (s OsStore) Get(ctx context.Context, d Digest) (m Meta, err error) {
	_, m, err = s.open(ctx, d)
	return
}

func (s OsStore) Open(ctx context.Context, d Digest) (io.ReadSeekCloser, Meta, error) {
	p, m, err := s.open(ctx, d)
	if err != nil {
		return nil, m, err
	}

	f, err := os.Open(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// The file may be removed after the stat.
			err = ErrNotExist
		}
		return nil, m, fmt.Errorf("open blob: %w", err)
	}

	return f, m, nil
}

func (s OsStore) Label(ctx context.Context, d Digest, labels Labels) error {
	return nil
}

func (s OsStore) Erase(ctx context.Context, d Digest) error {
	p := s.pathToRepo(d)
	if err := os.RemoveAll(p); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (s OsStore) open(_ context.Context, d Digest) (pb string, m Meta, err error) {
	pb = s.pathToRepo(d, "blob")
	pl := s.pathToRepo(d, "labels")

	info, err := os.Stat(pb)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = ErrNotExist
		} else {
			err = fmt.Errorf("stat: %w", err)
		}
		return
	}

	m.Digest = d
	m.Size = info.Size()

	var lf *os.File
	if lf, err = os.Open(pl); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			err = fmt.Errorf("open labels: %w", err)
			return
		}
	} else {
		defer lf.Close()

		m.Labels, err = readLabels(lf)
		if err != nil {
			err = fmt.Errorf("read labels: %w", err)
			return
		}
	}

	return
}

// checkDup checks if the blob with the given path already exists.
// It returns [ErrAlreadyExists] if it exists.
func (s OsStore) checkDup(p string) error {
	if _, err := os.Stat(p); err == nil {
		return ErrAlreadyExists
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat: %w", err)
	}
	return nil
}

func (s OsStore) pathToBlob(d Digest) string {
	v := string(d)
	return filepath.Join(s.root, "blobs", v[0:2], v[2:4], v[4:])
}

func (s OsStore) pathToRepo(d Digest, elem ...string) string {
	v := string(d)
	parts := make([]string, 0, 4+len(elem))
	parts = append(parts, s.repo, v[0:2], v[2:4], v[4:])
	parts = append(parts, elem...)
	return filepath.Join(parts...)
}

func (s OsStore) lock(ctx context.Context, d Digest) (func() error, error) {
	pl := filepath.Join(s.root, "locks", string(d))
	if err := os.MkdirAll(filepath.Dir(pl), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir locks: %w", err)
	}

	fl := flock.New(pl)
	for {
		if ok, err := fl.TryLockContext(ctx, 100*time.Millisecond); err != nil {
			return nil, err
		} else if ok {
			break
		}
	}

	return fl.Unlock, nil
}
