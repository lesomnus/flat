package flat

import (
	"bufio"
	"io"
	"net/http"
	"net/textproto"
)

type Labels = http.Header

func readLabels(r io.Reader) (Labels, error) {
	tp := textproto.NewReader(bufio.NewReader(r))

	h, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}
	return Labels(h), nil
}

func writeLabels(w io.Writer, labels Labels) error {
	return labels.Write(w)
}
