package gobls

import (
	"bufio"
	"io"
)

type scanner struct {
	br          *bufio.Reader
	buf         []byte // points within br's buffer
	longLineBuf []byte // our accumulation for long lines
	err         error
}

// NewScanner returns a scanner that reads from the specified `io.Reader`.  It
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
	return &scanner{
		br:          bufio.NewReader(r),
		longLineBuf: make([]byte, 0, DefaultBufferSize),
	}
}

// Bytes returns the byte slice that was just scanned.
func (s scanner) Bytes() []byte {
	return s.buf
}

// Err returns the error object associated with this scanner, or nil if no
// errors have occurred.
func (s scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

// Scan will scan the text from the `io.Reader`, and return true if scanning
// ought to continue or false if scanning is complete, because of error or
// EOF.
func (s *scanner) Scan() bool {
	s.buf, s.err = s.br.ReadSlice('\n')
	if s.err == nil {
		if l := len(s.buf); l > 0 && s.buf[l-1] == '\n' {
			if l > 1 && s.buf[l-2] == '\r' {
				s.buf = s.buf[:l-2]
			} else {
				s.buf = s.buf[:l-1]
			}
		}
		return true
	}
	if s.err == io.EOF {
		l := len(s.buf)
		more := l > 0
		if more && s.buf[l-1] == '\r' {
			s.buf = s.buf[:l-1]
		}
		return more
	}
	if s.err != bufio.ErrBufferFull {
		return false
	}

	s.longLineBuf = append(s.longLineBuf[:0], s.buf...) // init then copy volatile bytes from bufio

readMore:
	s.buf, s.err = s.br.ReadSlice('\n')
	s.longLineBuf = append(s.longLineBuf, s.buf...) // copy volatile bytes from bufio

	if s.err == nil {
		s.buf = s.longLineBuf
		if l := len(s.buf); l > 0 && s.buf[l-1] == '\n' {
			if l > 1 && s.buf[l-2] == '\r' {
				s.buf = s.buf[:l-2]
			} else {
				s.buf = s.buf[:l-1]
			}
		}
		return true
	}
	if s.err == io.EOF {
		l := len(s.buf)
		more := l > 0
		if more && s.buf[l-1] == '\r' {
			s.buf = s.buf[:l-1]
		}
		return more
	}
	if s.err == bufio.ErrBufferFull {
		goto readMore
	}
	return false
}

// String returns the string representation of the byte slice returned by the
// most recent Scan call.  DEPRECATED: Use the Text method.
func (s scanner) String() string {
	return string(s.buf)
}

// Text returns the string representation of the byte slice returned by the most
// recent Scan call.
func (s scanner) Text() string {
	return string(s.buf)
}
