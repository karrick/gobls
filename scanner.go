package gobls

import (
	"bufio"
	"io"
)

type scanner struct {
	br  *bufio.Reader
	bs  []byte
	buf []byte
	err error
}

// NewScanner returns a scanner that reads from the specified `io.Reader`.  It allocates a scanning
// buffer with the default buffer size.  This per-scanner buffer will grow to accomodate extremely
// long lines.
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
		br:  bufio.NewReader(r),
		buf: make([]byte, 0, DefaultBufferSize),
	}
}

// Bytes returns the byte slice that was just scanned.
func (s scanner) Bytes() []byte {
	return s.bs
}

// Err returns the error object associated with this scanner, or nil
// if no errors have occurred.
func (s scanner) Err() error {
	return s.err
}

// Scan will scan the text from the `io.Reader`, and return true if
// scanning ought to continue or false if scanning is complete,
// because of error or EOF. If true
func (s *scanner) Scan() bool {
	var isPrefix bool
	s.bs, isPrefix, s.err = s.br.ReadLine()
	if s.err != nil {
		if s.err == io.EOF {
			s.err = nil
		}
		return false
	}
	if !isPrefix {
		return true
	}

	// found a long line
	s.buf = append(s.buf[:0], s.bs...)
	for {
		s.bs, isPrefix, s.err = s.br.ReadLine()
		s.buf = append(s.buf, s.bs...)
		if s.err != nil {
			if s.err == io.EOF {
				s.err = nil
			}
			s.bs = s.buf
			return false
		}
		if !isPrefix {
			s.bs = s.buf
			return true
		}
	}
}

// String returns the string representation of the byte slice returned
// by the most recent Scan call.  DEPRECATED:  Use the Text method.
func (s scanner) String() string {
	return string(s.bs)
}

// Text returns the string representation of the byte slice returned
// by the most recent Scan call.
func (s scanner) Text() string {
	return string(s.bs)
}
