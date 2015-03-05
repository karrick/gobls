package gobls

import (
	"bufio"
	"bytes"
	"io"
)

type Scanner interface {
	Bytes() []byte
	Err() error
	Scan() bool
	String() string
}

type scanner struct {
	br  *bufio.Reader
	bs  []byte
	err error
}

func NewScanner(r io.Reader) Scanner {
	return &scanner{br: bufio.NewReader(r)}
}

func (s scanner) Bytes() []byte {
	return s.bs
}

func (s scanner) Err() error {
	return s.err
}

func (s scanner) String() string {
	return string(s.bs)
}

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
	buf := bytes.NewBuffer(s.bs)
	for {
		s.bs, isPrefix, s.err = s.br.ReadLine()
		_, werr := buf.Write(s.bs)
		if s.err != nil {
			if s.err == io.EOF {
				s.err = nil
			}
			return false
		}
		if werr != nil {
			s.err = werr
			return false
		}
		if !isPrefix {
			s.bs = buf.Bytes()
			return true
		}
	}
}
