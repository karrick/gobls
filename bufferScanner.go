package gobls

import "bytes"

// BufferScanner enumerates newline terminated strings from a provided slice of
// bytes faster than bufio.Scanner and gobls.Scanner. This is particular useful
// when a program already has the entire buffer in a slice of bytes. This
// structure uses newline as the line terminator, but returns nether the newline
// nor an optional carriage return from each discovered string.
type BufferScanner struct {
	buf         []byte
	left, right int
	done, cr    bool
}

// NewBufferScanner returns a BufferScanner that enumerates newline terminated
// strings from buf.
func NewBufferScanner(buf []byte) Scanner {
	l := len(buf)
	if l == 0 {
		return &BufferScanner{done: true}
	}

	// Inspect the final byte for newline.
	l--
	if buf[l] != '\n' {
		return &BufferScanner{buf: buf}
	}

	// When buffer ends with newline, remove it, to simplify logic executed for
	// each loop.
	return &BufferScanner{buf: buf[:l]}
}

// Bytes returns the byte slice that was just scanned. It does not return the
// terminating newline character, nor any optional preceding carriage return
// character.
func (b *BufferScanner) Bytes() []byte {
	if b.cr {
		return b.buf[b.left : b.right-1]
	}
	return b.buf[b.left:b.right]
}

// Err returns nil because scanning from a slice of bytes will never cause an
// error.
func (b *BufferScanner) Err() error { return nil }

// Scan will scan the text from the original slice of bytes, and return true if
// scanning ought to continue or false if scanning is complete, because of the
// end of the slice of bytes.
func (b *BufferScanner) Scan() bool {
	if b.done {
		b.buf = nil
		b.cr = false
		b.left = 0
		b.right = 0
		return false
	}

	if b.right > 0 {
		// Trim previous line.
		b.left = b.right + 1
	}

	next := bytes.IndexByte(b.buf[b.left:], '\n')

	if next == -1 {
		b.done = true
		b.right = len(b.buf)
	} else {
		b.right = b.left + next
	}

	// Is the final character a carriage return?
	b.cr = b.right > 0 && b.buf[b.right-1] == '\r'

	return true
}

// Text returns the string representation of the byte slice returned by the most
// recent Scan call. It does not return the terminating newline character, nor
// any optional preceding carriage return character.
func (b *BufferScanner) Text() string { return string(b.Bytes()) }
