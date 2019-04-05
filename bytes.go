package gobls

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"

	"github.com/karrick/gorill"
)

// Results represents an iteration cursor into the results from a range query.
type Results struct {
	buf   []byte
	entry []byte // only needed for Scan
}

// newResultsFromReader returns a Results structure after reading the provided
// io.Reader, or an error if either reading resulted in an error.  This
// initialization function is provided because Results will not properly work if
// the final byte read is not a newline.
func newResultsFromReader(r io.Reader) (*Results, error) {
	buf, err := ioutil.ReadAll(&gorill.LineTerminatedReader{R: r})
	if err != nil {
		return nil, err
	}
	return &Results{buf: buf}, nil
}

func (r *Results) Bytes() []byte { return r.buf }

func (r *Results) Split() []string {
	if len(r.buf) == 1 {
		// Because we only create Results from buffers that end with newlines,
		// when we have a single byte in the buffer, it must be the newline,
		// which means we have no results.
		return nil
	}
	slice := strings.Split(string(r.buf), "\n")
	return slice[:len(slice)-1] // remove the final empty element
}

// Scan advances the results, returning true if and only if at least one more
// entry is available.
//
// When the caller intends to visit each result string, then Split is about
// eight times as performant as Scan.  However, Scan is much more performant
// when only a small subset of results is desired.
func (r *Results) Scan() bool {
	if r.buf == nil {
		return false
	}
	i := bytes.IndexByte(r.buf, '\n')
	// fmt.Fprintf(os.Stderr, "buf: %q; i: %d\n", r.buf, i)
	switch i {
	case -1, 0:
		r.entry = nil
		r.buf = nil
		return false
	default:
		r.entry = r.buf[:i]
		r.buf = r.buf[i+1:]
		// fmt.Fprintf(os.Stderr, "  entry: %q; buf: %q\n", r.entry, r.buf)
	}
	return true
}

// Text returns the string corresponding to the current entry in the results.
func (r *Results) Text() string {
	return string(r.entry)
}
