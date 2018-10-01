package gobls

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

const minRead = 1024 // must be smaller than DefaultBufferSize

func debug(format string, a ...interface{}) {
	if false {
		fmt.Fprintf(os.Stderr, format+"\n", a...)
	}
}

type scanner struct {
	source io.Reader
	//            l    n    r
	// rbuf: 01234567890123456789
	//
	rbuf     []byte // rbuf stores bytes read from source
	l        int    // left most index for searching for next newline
	i        int    // index of most recently found newline
	r        int    // right most index for searching for next newline
	width    int    // width of previous CR?LF combination
	llbuf    []byte // long line buffer stores next long line
	overflow bool   // true when line is concat between llbuf and rbuf[:s.l]
	err      error
}

// NewScanner returns a scanner that reads from the specified `io.Reader`. It
// allocates a scanning buffer with the default buffer size.  This per-scanner
// buffer will grow to accomodate extremely long lines.
//
//    var lines, characters int
//    ls := gobls.NewScanner(os.Stdin)
//    for ls.Scan() {
//        lines++
//        characters += len(ls.Bytes())
//    }
//    if ls.Err() != nil {
//        fmt.Fprintln(os.Stderr, "cannot scan:", ls.Err())
//    }
//    fmt.Println("Counted",lines,"and",characters,"characters.")
func NewScanner(r io.Reader) Scanner {
	return NewScannerSize(r, DefaultBufferSize)
}

// NewScannerSize returns a scanner that reads from the specified
// `io.Reader`. It allocates a scanning buffer with the specified buffer
// size. This per-scanner buffer will grow to accomodate extremely long lines.
//
//    var lines, characters int
//    ls := gobls.NewScannerSize(os.Stdin, 4096)
//    for ls.Scan() {
//        lines++
//        characters += len(ls.Bytes())
//    }
//    if ls.Err() != nil {
//        fmt.Fprintln(os.Stderr, "cannot scan:", ls.Err())
//    }
//    fmt.Println("Counted",lines,"and",characters,"characters.")
func NewScannerSize(r io.Reader, size int) Scanner {
	return &scanner{
		source: r,
		rbuf:   make([]byte, size),    // read buffer with len
		llbuf:  make([]byte, 0, size), // line buffer with 0 len
	}
}

// Bytes returns the byte slice that was just scanned.
func (s *scanner) Bytes() []byte {
	if s.overflow {
		s.overflow = false
		return append(s.llbuf, s.rbuf[s.l:s.i]...)
	}

	return s.rbuf[s.l:s.i]
}

// Err returns the error object associated with this scanner, or nil if no
// errors have occurred.
func (s scanner) Err() error {
	return s.err
}

// Scan will scan the text from the `io.Reader`, and return true if scanning
// ought to continue or false if scanning is complete, because of error or
// EOF. If true
func (s *scanner) Scan() bool {
	// Each search ought start at previous EOL plus its width, but keep in
	// mind s.n could have matched at s.r, so the below could make s.l larger
	// than s.r.
	if s.width > 0 {
		s.i += s.width
		s.width = 0 // once increment i by width, then zero width
	}
	s.l = s.i

	// Continue reading and searching until either a read error, or a newline
	// has been found.
	for {
		if s.i >= s.r {
			// debug("need to read more bytes")

			if len(s.rbuf)-s.r < minRead {
				// debug("but not enough room for read")

				if false && s.l >= minRead {
					// We can read more after we shift over by s.l bytes.
					nc := copy(s.rbuf[:s.r-s.l], s.rbuf[s.l:s.r])
					// debug("Shifted %d bytes to the left by %d bytes", nc, s.l)
					s.i -= s.l
					s.l = 0
					s.r = nc
				} else {
					// Copy read buffer contents to overflow buffer.
					if !s.overflow {
						s.overflow = true
						s.llbuf = s.llbuf[:0] // reduce GC when resetting
					}
					s.llbuf = append(s.llbuf, s.rbuf[s.l:s.r]...)
					s.i = 0
					s.l = 0
					s.r = 0
				}
			}

			// debug("Read at least one byte or error")
			var nr int
			for nr == 0 {
				// Robust handling of Read as described by io.Reader
				// documentation says to interpret read of 0 bytes and nil
				// error as nothing happened. In this case, there's nothing we
				// can do, until we get at least one byte. Therefore, go ahead
				// and read again.
				nr, s.err = s.source.Read(s.rbuf[s.r:])
				// debug("Read %d bytes, %s error", nr, s.err)
				s.r += nr

				if s.err != nil {
					if s.err == io.EOF {
						s.err = nil
						s.i = s.r // Additional Bytes check ought return up to the final rune consumed.
						if s.i > s.l {
							return true
						}
					}
					return false
				}
			}
		}

		// debug("Search L: %d; I: %d; W: %d; R: %d", s.l, s.i, s.width, s.r)

		// Treat the byte sequence as UTF-8 string.
		index, width := indexEOL(s.rbuf[s.i:s.r])
		// debug("scan returned %d, %d", index, width)
		if index >= 0 {
			s.i += index // mark EOL
			s.width = width
			return true
		}

		// // debug("cannot find newline in %q", string(s.rbuf[s.i:s.r]))
		s.i = s.r // do not need to scan those runes again
	}
}

// indexEOL returns the index of CR?LF and the number of bytes used by EOL, or
// -1 and 0, if CR?LF was not found.
func indexEOL(buf []byte) (int, int) {
	index := bytes.IndexRune(buf, '\n')
	if index == -1 {
		return -1, 0
	}

	// Decode last rune from buf with LF.* removed.
	r, _ := utf8.DecodeLastRune(buf[:index])
	switch r {
	case '\r':
		// CRLF takes 2 bytes
		return index - 1, 2
	default:
		// When find any other rune, including utf8.RuneError, just report width
		// of 1 to account for the LF.
		return index, 1
	}
}

func indexEOLQuick(buf []byte) (int, int) {
	// byte slice represents UTF-8 string, so prefer to use bytes.IndexRune()
	// to march through bytes as a sequence of runes, but investigating
	// whether this is more fast.
	switch index := bytes.IndexByte(buf, '\n'); index {
	case -1:
		return -1, 0
	case 0:
		return index, 1
	default:
		if i := index - 1; buf[i] == '\r' {
			return i, 2
		}
		return index, 1
	}
}

func indexEOL1(buf []byte) (int, int) {
	// I'm not entirely confident in this particular approach, because it may be
	// possible for a multibyte rune to have one of its bytes equal to 0x10.
	index := bytes.IndexByte(buf, '\n')
	if index == -1 {
		return -1, 0
	}

	// Decode last rune from buf with LF.* removed.
	r, _ := utf8.DecodeLastRune(buf[:index])
	switch r {
	case '\r':
		// CRLF takes 2 bytes
		return index - 1, 2
	default:
		// When find any other rune, including utf8.RuneError, just report width
		// of 1 to account for the LF.
		return index, 1
	}
}

// String returns the string representation of the byte slice returned by the
// most recent Scan call.  DEPRECATED: Use the Text method.
func (s *scanner) String() string {
	return string(s.Bytes())
}

// Text returns the string representation of the byte slice returned by the most
// recent Scan call.
func (s *scanner) Text() string {
	return string(s.Bytes())
}
