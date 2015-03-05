package gobls

import (
	"bufio"
	"bytes"
	"io"
)

type Scanner interface {
	Scan() ([]byte, error)
}

type scanner struct {
	br *bufio.Reader
}

func NewScanner(r io.Reader) Scanner {
	return &scanner{br: bufio.NewReader(r)}
}

func (s *scanner) Scan() ([]byte, error) {
	line, isPrefix, rerr := s.br.ReadLine()
	if rerr != nil {
		return line, rerr
	}
	if !isPrefix {
		return line, nil
	}
	// here's a long line
	buf := bytes.NewBuffer(line)
	for {
		line, isPrefix, rerr = s.br.ReadLine()
		_, werr := buf.Write(line)
		if rerr != nil {
			return buf.Bytes(), rerr
		}
		if werr != nil {
			return buf.Bytes(), werr
		}
		if !isPrefix {
			return buf.Bytes(), nil
		}
	}
}
