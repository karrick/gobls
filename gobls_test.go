package gobls

import (
	"bufio"
	"bytes"
	"testing"
)

const (
	lineCount  = 100
	lineLength = 1024
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

func TestVeryLargeLinesRequireSingleInvocation(t *testing.T) {
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

func BenchmarkBufioScanner(b *testing.B) {
	master := makeBuffer(lineCount, lineLength)
	initial := master.Bytes()
	var line string
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := bufio.NewScanner(bytes.NewBuffer(initial))
		for s.Scan() {
			line = s.Text()
		}
	}
	if len(line) != lineLength {
		b.Errorf("Actual: %#v; Expected: %#v", len(line), lineLength)
	}
}

func BenchmarkGobls(b *testing.B) {
	master := makeBuffer(lineCount, lineLength)
	initial := master.Bytes()
	var line string
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := NewScanner(bytes.NewBuffer(initial))
		for s.Scan() {
			line = s.String()
		}
	}
	if len(line) != lineLength {
		b.Errorf("Actual: %#v; Expected: %#v", len(line), lineLength)
	}
}
