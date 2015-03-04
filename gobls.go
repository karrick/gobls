package gobls

import (
	"bufio"
	"bytes"
	"io"
)

type Scanner interface {
	Scan() (string, error)
}

type scanner struct {
	br *bufio.Reader
}

func NewScanner(r io.Reader) Scanner {
	return &scanner{br: bufio.NewReader(r)}
}

func (s *scanner) Scan() (string, error) {
	line, isPrefix, err := s.br.ReadLine()
	if err != nil {
		return string(line), err
	}
	if !isPrefix {
		return string(line), nil
	}
	// here's a long line
	buf := bytes.NewBuffer(line)
	for {
		line, isPrefix, rerr := s.br.ReadLine()
		_, werr := buf.Write(line)
		if rerr != nil {
			return buf.String(), rerr
		}
		if werr != nil {
			return buf.String(), werr
		}
		if !isPrefix {
			return buf.String(), nil
		}
	}
}
