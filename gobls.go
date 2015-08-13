package gobls

import (
	"bufio"
	"io"
	"sync"
)

// pool is a goroutine safe pool of very large buffers to use for long
// lines.
var pool sync.Pool

func init() {
	pool.New = func() interface{} {
		return make([]byte, 0, bufio.MaxScanTokenSize)
	}
}

// Scanner provides an interface for reading newline-delimited lines
// of text. It is similar to `bufio.Scanner`, but wraps
// `bufio.ReadLine` so lines of arbitrary length can be
// scanned. Successive calls to the Scan method will step through the
// lines of a file, skipping the newline whitespace between lines.
//
// Scanning stops unrecoverably at EOF, or at the first I/O
// error. Unlike `bufio.Scanner`, howver, attempting to scan a line
// longer than `bufio.MaxScanTokenSize` will not result in an error,
// but will return the long line.
//
// It is not necessary to check for errors by calling the Err method
// until after scanning stops, when the Scan method returns false.
type Scanner interface {
	Bytes() []byte
	Err() error
	Scan() bool
	Text() string
	String() string
}

type scanner struct {
	br  *bufio.Reader
	bs  []byte
	err error
}

// NewScanner returns a scanner that reads from the specified
// `io.Reader`.
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
	return &scanner{br: bufio.NewReader(r)}
}

// NewScanner returns a scanner that reads from the specified
// `io.Reader`, using internal buffers with at least the specified
// size number of bytes. This is useful when you know the most common
// line length.
func NewScannerSize(r io.Reader, size int) Scanner {
	return &scanner{br: bufio.NewReaderSize(r, size)}
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

	// here's a long line
	buf := pool.Get().([]byte)
	buf = append(buf[:0], s.bs...)
	defer func() {
		// need to copy data out before we return buf to pool
		s.bs = append(make([]byte, 0, len(buf)), buf...)
		pool.Put(buf)
	}()
	for {
		s.bs, isPrefix, s.err = s.br.ReadLine()
		buf = append(buf, s.bs...)
		if s.err != nil {
			if s.err == io.EOF {
				s.err = nil
			}
			return false
		}
		if !isPrefix {
			return true
		}
	}
}

// String returns the string representation of the byte slice returned
// by the most recent Scan call.
func (s scanner) String() string {
	return string(s.bs)
}

// Text returns the string representation of the byte slice returned
// by the most recent Scan call.
func (s scanner) Text() string {
	return string(s.bs)
}
