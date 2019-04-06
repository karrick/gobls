package gobls

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
)

// response represents an iteration cursor into the results from a range query.
type response struct {
	buf   []byte
	entry []byte // only needed for Scan
}

// newResponseFromReader returns a response instance after reading the provided
// io.Reader, or an error if reading resulted in an error.  This initialization
// function is provided because some range server implementations return a final
// newline and some do not.  This function normalizes those response.
func newResponseFromReader(r io.Reader) (*response, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	// When final byte is newline, trim it.
	if l := len(buf); l > 0 && buf[l-1] == '\n' {
		buf = buf[:l-1]
	}
	return &response{buf: buf}, nil
}

// Split returns a slice of strings, each string representing one line from the
// response.
func (r *response) Split() []string {
	if len(r.buf) > 0 {
		return strings.Split(string(r.buf), "\n")
	}
	return nil
}

// Scan advances the results, returning true if and only if at least one more
// entry is available.
//
// When the caller intends to visit each result string, then Split is about
// eight times as performant as Scan.  However, Scan is much more performant
// when only a small subset of results is desired.
func (r *response) Scan() bool {
	if len(r.buf) == 0 {
		return false
	}
    if i := bytes.IndexByte(r.buf, '\n'); i >= 0 {
		r.entry = r.buf[:i]
		r.buf = r.buf[i+1:]
		return true
    }
	r.entry = r.buf
	r.buf = nil
	return true
}

// Text returns the string corresponding to the current entry in the results.
func (r *response) Text() string {
	return string(r.entry)
}
