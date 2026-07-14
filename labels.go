package flob

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"net/textproto"
	"slices"
)

type Labels = http.Header

// cloneLabels returns a deep copy of l: both the map and each value slice are copied, so the
// result shares no backing storage with l. maps.Clone alone would copy the map but alias the
// underlying []string values, letting a caller's mutation of a returned label leak back into a
// store's in-memory copy.
func cloneLabels(l Labels) Labels {
	if l == nil {
		return nil
	}
	c := make(Labels, len(l))
	for k, vs := range l {
		c[k] = slices.Clone(vs)
	}
	return c
}

func readLabels(r io.Reader) (Labels, error) {
	tp := textproto.NewReader(bufio.NewReader(r))

	h, err := tp.ReadMIMEHeader()
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return Labels(h), nil
}

func writeLabels(w io.Writer, labels Labels) error {
	return labels.Write(w)
}
