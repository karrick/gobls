package gobls

import (
	"bufio"
	"bytes"
	"testing"
)

const (
	lineCount          = 100
	shortLineLength    = 100
	avgLineLength      = 1024
	longLineLength     = 4096 - 2
	veryLongLineLength = bufio.MaxScanTokenSize - 2
)

func makeBuffer(lineCount, lineLength int) *bytes.Buffer {
	buf := make([]byte, 0, lineCount*(lineLength+2))
	bb := bytes.NewBuffer(buf)
	for line := 0; line < lineCount; line++ {
		for i := 0; i < lineLength; i++ {
			bb.WriteByte('a')
		}
		bb.WriteString("\r\n")
	}
	return bb
}

func TestNoEOF(t *testing.T) {
	bb := bytes.NewBufferString("")
	s := NewScanner(bb)
	for s.Scan() {
		t.Errorf("Actual: scan returned true; Expected: false")
	}
	if s.Err() != nil {
		t.Errorf("Actual: %#v; Expected: %#v", s.Err(), nil)
	}
}

func TestSequencesThroughEntireBuffer(t *testing.T) {
	bb := bytes.NewBufferString("flubber\nblubber\nfoo")
	s := NewScanner(bb)
	expectedLines := []string{"flubber", "blubber", "foo"}
	actualLines := make([]string, 0)
	for s.Scan() {
		actualLines = append(actualLines, s.String())
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

func TestLongLinesRequireSingleInvocation(t *testing.T) {
	r := makeBuffer(1, longLineLength)
	line := r.String()
	line = line[:len(line)-2] // trim CRLF

	s := NewScanner(r)
	lines := make([]string, 0, 1)
	for s.Scan() {
		lines = append(lines, s.String())
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

func TestVeryLongLinesRequireSingleInvocation(t *testing.T) {
	r := makeBuffer(1, bufio.MaxScanTokenSize+5)
	line := r.String()
	line = line[:len(line)-2] // trim CRLF

	s := NewScanner(r)
	lines := make([]string, 0, 1)
	for s.Scan() {
		lines = append(lines, s.String())
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

type simpleScanner interface {
	Err() error
	Scan() bool
	Bytes() []byte
}

func benchmarkScanner(b *testing.B, lineLength int, makeScanner func(*bytes.Buffer) simpleScanner) {
	master := makeBuffer(lineCount, lineLength)
	initial := master.Bytes()
	var line []byte
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := makeScanner(bytes.NewBuffer(initial))
		for s.Scan() {
			line = s.Bytes()
		}
		if s.Err() != nil {
			b.Fatal("Actual: %#v; Expected: %#v", s.Err(), nil)
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

func BenchmarkGoblsScannerVeryLongEx(b *testing.B) {
	makeScanner := func(bb *bytes.Buffer) simpleScanner {
		return NewScannerSize(bb, bufio.MaxScanTokenSize)
	}
	benchmarkScanner(b, veryLongLineLength, makeScanner)
}
