package gobls

import "bytes"

// bufferScanner enumerates newline terminated strings from a provided slice of
// bytes with minimal heap allocations.
type bufferScanner struct {
	buf      []byte
	right    int
	done, cr bool
}

// newBufferScanner returns a bufferScanner that enumerates newline terminated
// strings from buf.
func newBufferScanner(buf []byte) Scanner {
	// debug("NEW: %q\n", buf)
	l := len(buf)
	if l == 0 {
		return &bufferScanner{done: true}
	}

	// Inspect the final byte.
	l--
	if buf[l] != '\n' {
		return &bufferScanner{buf: buf}
	}

	// When buffer ends with newline, remove it, to simplify logic executed for
	// each loop.
	// debug("NEW: removing newline: %q -> %q\n", buf, buf[:l])
	return &bufferScanner{buf: buf[:l]}
}

// Bytes returns the byte slice that was just scanned.
func (b *bufferScanner) Bytes() []byte {
	if b.cr {
		return b.buf[:b.right-1]
	}
	return b.buf[:b.right]
}

// Err returns nil because scanning from a slice of bytes will never cause an
// error.
func (b *bufferScanner) Err() error { return nil }

// Scan will scan the text from the original slice of bytes, and return true if
// scanning ought to continue or false if scanning is complete, because of error
// or EOF.
func (b *bufferScanner) Scan() bool {
	// debug("SCAN: %q (%d)\n", b.buf, b.right)
	if b.done {
		// debug("SCAN: we were already done\n")
		b.buf = nil
		b.cr = false
		b.right = 0
		return false
	}

	if b.right > 0 {
		// Trim previous line.
		// debug("SCAN: before trim: %q\n", b.buf)
		b.buf = b.buf[b.right+1:]
		// debug("SCAN: after  trim: %q\n", b.buf)
	}

	next := bytes.IndexRune(b.buf, '\n')
	// debug("SCAN: buf: %q; index right: %d\n", b.buf, b.right)
	b.right = next

	if b.right == -1 {
		b.done = true
		b.right = len(b.buf)
	}
	b.cr = b.right > 0 && b.buf[b.right-1] == '\r'

	// debug("SCAN: results: %q; cr: %t\n", b.buf[:b.right], b.cr)
	return true
}

// Text returns the string representation of the byte slice returned by the most
// recent Scan call.
func (b *bufferScanner) Text() string {
	return string(b.Bytes())
}
