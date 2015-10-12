package gobls

import (
	"bufio"
	"bytes"
	"testing"
)

const (
	lineCount          = 1000
	shortLineLength    = 100
	avgLineLength      = 1024
	longLineLength     = 4096 - 2
	veryLongLineLength = bufio.MaxScanTokenSize - 2
)

func makeBytes(lineCount, lineLength int) []byte {
	buf := make([]byte, 0, lineCount*(lineLength+2))
	for line := 0; line < lineCount; line++ {
		for i := 0; i < lineLength; i++ {
			buf = append(buf, 'a')
		}
		buf = append(buf, '\r', '\n')
	}
	return buf
}

func TestNoEOF(t *testing.T) {
	test := func(s simpleScanner) {
		for s.Scan() {
			t.Errorf("Actual: scan returned true; Expected: false")
		}
		if s.Err() != nil {
			t.Errorf("Actual: %#v; Expected: %#v", s.Err(), nil)
		}
	}

	test(NewScanner(bytes.NewBufferString("")))
	test(bufio.NewScanner(bytes.NewBufferString("")))
}

func TestSequencesThroughEntireBuffer(t *testing.T) {
	test := func(s simpleScanner) {
		expectedLines := []string{"flubber", "blubber", "foo"}
		actualLines := make([]string, 0)
		for s.Scan() {
			actualLines = append(actualLines, s.Text())
		}
		if s.Err() != nil {
			t.Errorf("Actual: %#v; Expected: %#v", s.Err(), nil)
		}
		if len(actualLines) != len(expectedLines) {
			t.Fatalf("Actual: %#v; Expected: %#v", len(actualLines), len(expectedLines))
		}
		for i := 0; i < len(expectedLines); i++ {
			if actualLines[i] != expectedLines[i] {
				t.Errorf("Actual: %#v; Expected: %#v",
					actualLines[i], expectedLines[i])
			}
		}
	}

	test(NewScanner(bytes.NewBufferString("flubber\nblubber\nfoo")))
	test(bufio.NewScanner(bytes.NewBufferString("flubber\nblubber\nfoo")))
}

func TestLongLinesRequireSingleInvocation(t *testing.T) {
	buf := makeBytes(1, longLineLength)
	line := string(buf)
	line = line[:len(line)-2] // trim CRLF

	test := func(s simpleScanner) {
		lines := make([]string, 0, 1)
		for s.Scan() {
			lines = append(lines, s.Text())
		}
		if s.Err() != nil {
			t.Errorf("Actual: %#v; Expected: %#v", s.Err(), nil)
		}
		if len(lines) != 1 {
			t.Fatalf("Actual: %#v; Expected: %#v", len(lines), 1)
		}
		if lines[0] != line {
			t.Errorf("Actual: %#v; Expected: %#v", lines[0], line)
		}
	}

	test(NewScanner(bytes.NewBuffer(buf)))
	test(bufio.NewScanner(bytes.NewBuffer(buf)))
}

func TestVeryLongLinesRequireSingleInvocation(t *testing.T) {
	buf := makeBytes(1, bufio.MaxScanTokenSize+5)
	line := string(buf)
	line = line[:len(line)-2] // trim CRLF

	test := func(s simpleScanner) {
		lines := make([]string, 0, 1)
		for s.Scan() {
			lines = append(lines, s.Text())
		}
		if s.Err() != nil {
			t.Errorf("Actual: %#v; Expected: %#v", s.Err(), nil)
		}
		if len(lines) != 1 {
			t.Fatalf("Actual: %#v; Expected: %#v", len(lines), 1)
		}
		if lines[0] != line {
			t.Errorf("Actual: %#v; Expected: %#v", lines[0], line)
		}
	}

	test(NewScanner(bytes.NewBuffer(buf)))
	// test(bufio.NewScanner(bytes.NewBuffer(buf))) // err = token too long
}

type simpleScanner interface {
	Err() error
	Scan() bool
	Bytes() []byte
	Text() string
}

func benchmarkScanner(b *testing.B, lineLength int, makeScanner func(*bytes.Buffer) simpleScanner) {
	master := makeBytes(lineCount, lineLength)
	var line []byte
	s := makeScanner(bytes.NewBuffer(master))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for s.Scan() {
			line = s.Bytes()
		}
		if err := s.Err(); err != nil {
			b.Fatalf("Actual: %#v; Expected: %#v", err, nil)
		}
	}
	if len(line) != lineLength {
		b.Errorf("Actual: %#v; Expected: %#v", len(line), lineLength)
	}
}

func BenchmarkBufioScannerAverage(b *testing.B) {
	makeScanner := func(bb *bytes.Buffer) simpleScanner {
		return bufio.NewScanner(bb)
	}
	benchmarkScanner(b, avgLineLength, makeScanner)
}

func BenchmarkGoblsScannerAverage(b *testing.B) {
	makeScanner := func(bb *bytes.Buffer) simpleScanner {
		return NewScanner(bb)
	}
	benchmarkScanner(b, avgLineLength, makeScanner)
}

func BenchmarkBufioScannerShort(b *testing.B) {
	makeScanner := func(bb *bytes.Buffer) simpleScanner {
		return bufio.NewScanner(bb)
	}
	benchmarkScanner(b, shortLineLength, makeScanner)
}

func BenchmarkGoblsScannerShort(b *testing.B) {
	makeScanner := func(bb *bytes.Buffer) simpleScanner {
		return NewScanner(bb)
	}
	benchmarkScanner(b, shortLineLength, makeScanner)
}

func BenchmarkBufioScannerLong(b *testing.B) {
	makeScanner := func(bb *bytes.Buffer) simpleScanner {
		return bufio.NewScanner(bb)
	}
	benchmarkScanner(b, longLineLength, makeScanner)
}

func BenchmarkGoblsScannerLong(b *testing.B) {
	makeScanner := func(bb *bytes.Buffer) simpleScanner {
		return NewScanner(bb)
	}
	benchmarkScanner(b, longLineLength, makeScanner)
}

func BenchmarkBufioScannerVeryLong(b *testing.B) {
	makeScanner := func(bb *bytes.Buffer) simpleScanner {
		return bufio.NewScanner(bb)
	}
	benchmarkScanner(b, veryLongLineLength, makeScanner)
}

func BenchmarkGoblsScannerVeryLong(b *testing.B) {
	makeScanner := func(bb *bytes.Buffer) simpleScanner {
		return NewScanner(bb)
	}
	benchmarkScanner(b, veryLongLineLength, makeScanner)
}
