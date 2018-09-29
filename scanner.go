package gobls

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
)

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
	i        int    // index where ought to start next search
	n        int    // index of most recently found newline
	r        int    // right most index for searching for next newline
	w        int    // width of previous CR?LF combination
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
		b := append(s.llbuf, s.rbuf[s.l:s.i]...)
		s.llbuf = s.llbuf[:0]
		return b
	}

	debug("Bytes L: %d; I: %d; W: %d; R: %d", s.l, s.i, s.w, s.r)
	b := s.rbuf[s.l:s.i]
	// debug("\t%q", string(b))

	return b
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
	const minRead = 512
	const minShift = 512

	// Note s.i is initialized to -1, so on first run this rolls over to
	// zero. Each search ought start after previous newline, but keep in mind
	// s.n could have matched at s.r, so the below could make s.l larger than
	// s.r.
	s.i += s.w
	s.l = s.i

	// Continue reading and searching until either a read error, or a newline
	// has been found.
	for {
		debug("Scan L: %d; I: %d; W: %d; R: %d", s.l, s.i, s.w, s.r)
		if false && s.l <= s.i && s.i <= s.r && s.r < len(s.rbuf) {
			debug("\t%q %q", s.rbuf[s.l:s.i], s.rbuf[s.i:s.r])
		}
		if s.i >= s.r {
			debug("we need to read more bytes")
			if len(s.rbuf)-s.r < minRead {
				debug("but there's not enough room to do a read")
				if s.l < minShift {
					debug("and there's not enough room to shift: Append rbuf to overflow, clear rbuf, and keep looking.")
					s.llbuf = append(s.llbuf, s.rbuf[s.l:s.r]...)
					s.i = 0
					s.l = 0
					s.n = 0
					s.r = 0
					s.overflow = true
					continue
				}

				debug("Shift what we have to the left")
				nc := copy(s.rbuf, s.rbuf[s.l:s.r])
				s.l -= nc
				s.i -= nc
				s.n -= nc
				s.r -= nc
				continue
			}

			var nr int
			for nr == 0 {
				debug("Read at least one byte or error")
				nr, s.err = s.source.Read(s.rbuf[s.r:])
				debug("Was able to read %d bytes", nr)
				s.r += nr

				// If read some bytes, then ignore error for now and continue.
				if nr == 0 && s.err != nil {
					debug("read error: %s", s.err)
					if s.err == io.EOF {
						s.err = nil
						s.i = s.r // additional Bytes check ought return up to final rune
						if s.i > s.l {
							return true
						}
					}
					return false
				}
			}
		}

		debug("Search L: %d; I: %d; W: %d; R: %d", s.l, s.i, s.w, s.r)

		// Treat the byte sequence as UTF-8 string. Would prefer to use
		// bytes.IndexRune(), that have to either scan for CR and check whether
		// following rune is LF, and if no CR then scan for LF; or scan for LF
		// and check whether previous rune is CR.
		index, width := indexEOL(s.rbuf[s.i:s.r])
		debug("scan returned %d, %d", index, width)
		if index >= 0 {
			s.i += index // never need to scan these bytes again
			s.w = width
			return true
		}

		// debug("cannot find newline in %q", string(s.rbuf[s.i:s.r]))
		s.i = s.r // do not need to scan those runes again
	}
}

// indexEOL returns the index of CR?LF and the number of bytes used, or -1 and
// 0, if CR?LF was not found.
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
		// Any other rune, including utf8.RuneError, just report width of 1.
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
