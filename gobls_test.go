package gobls

import (
	"bufio"
	"bytes"
	"io"
	"testing"
)

const (
	shortLineLength    = 100 - 2
	avgLineLength      = 1024 - 2
	longLineLength     = 4096 - 2
	veryLongLineLength = bufio.MaxScanTokenSize - 2
)

func makeBuffer(lineCount, lineLength int) []byte {
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
	test := func(s Scanner) {
		if actual, want := s.Scan(), false; actual != want {
			t.Errorf("Actual: %#v; Expected: %#v", actual, want)
		}
		if actual, want := s.Err(), error(nil); actual != want {
			t.Errorf("Actual: %#v; Expected: %#v", actual, want)
		}
	}

	corpus := ""
	test(bufio.NewScanner(bytes.NewBufferString(corpus)))
	test(NewScanner(bytes.NewBufferString(corpus)))
}

func TestSequencesThroughEntireBuffer(t *testing.T) {
	test := func(expected []string, s Scanner) {
		var actualLines []string
		for s.Scan() {
			actualLines = append(actualLines, s.Text())
		}
		if actual, want := s.Err(), error(nil); actual != want {
			t.Errorf("Actual: %#v; Expected: %#v", actual, want)
		}
		if actual, want := len(actualLines), len(expected); actual != want {
			t.Fatalf("Actual: %#v; Expected: %#v", actual, want)
		}
		for i := 0; i < len(expected); i++ {
			if actual, want := actualLines[i], expected[i]; actual != want {
				t.Errorf("Actual: %#v; Expected: %#v", actual, want)
			}
		}
	}

	corpus := "flubber\nblubber\nfoo"
	expected := []string{"flubber", "blubber", "foo"}
	test(expected, bufio.NewScanner(bytes.NewBufferString(corpus)))
	test(expected, NewScanner(bytes.NewBufferString(corpus)))
}

func TestLongLinesRequireSingleInvocation(t *testing.T) {
	buf := makeBuffer(1, longLineLength)
	line := string(buf)
	line = line[:len(line)-2] // trim CRLF

	test := func(s Scanner) {
		var lines []string
		for s.Scan() {
			lines = append(lines, s.Text())
		}
		if actual, want := s.Err(), error(nil); actual != want {
			t.Errorf("Actual: %#v; Expected: %#v", actual, want)
		}
		if actual, want := len(lines), 1; actual != want {
			t.Fatalf("Actual: %#v; Expected: %#v", actual, want)
		}
		if actual, want := lines[0], line; actual != want {
			t.Errorf("Actual: %#v; Expected: %#v", actual, want)
		}
	}

	test(bufio.NewScanner(bytes.NewReader(buf)))
	test(NewScanner(bytes.NewReader(buf)))
}

func TestVeryLongLinesRequireSingleInvocation(t *testing.T) {
	buf := makeBuffer(1, bufio.MaxScanTokenSize+5)
	line := string(buf)
	line = line[:len(line)-2] // trim CRLF

	test := func(s Scanner) {
		lines := make([]string, 0, 1)
		for s.Scan() {
			lines = append(lines, s.Text())
		}
		if actual, want := s.Err(), error(nil); actual != want {
			t.Errorf("Actual: %#v; Expected: %#v", actual, want)
		}
		if actual, want := len(lines), 1; actual != want {
			t.Fatalf("Actual: %#v; Expected: %#v", actual, want)
		}
		if actual, want := lines[0], line; actual != want {
			t.Errorf("Actual: %#v; Expected: %#v", actual, want)
		}
	}

	// test(bufio.NewScanner(bytes.NewReader(buf))) // bufio will return err = token too long
	test(NewScanner(bytes.NewReader(buf)))
}

func benchmarkScanner(b *testing.B, lineLength int, makeScanner func(io.Reader) Scanner) {
	wanted := makeBuffer(1, lineLength)
	wanted = wanted[:len(wanted)-2] // trim CRLF

	// NOTE: make buffer with line count set to b.N
	s := makeScanner(bytes.NewReader(makeBuffer(b.N, lineLength)))

	var line []byte
	var count int

	b.ResetTimer()
	for s.Scan() {
		line = s.Bytes()
		count++
	}

	if actual, want := s.Err(), error(nil); actual != want {
		b.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}
	// NOTE: ensure proper number of lines scanned
	if actual, want := count, b.N; actual != want {
		b.Errorf("Actual: %#v; Expected: %#v", actual, want)
	}
	// NOTE: test line contents to prevent compiler optimization from eliding call to s.Bytes()
	if !bytes.Equal(line, wanted) {
		b.Fatalf("Actual: %#v; Expected: %#v", line, wanted)
	}
}

func BenchmarkScannerAverageBufio(b *testing.B) {
	benchmarkScanner(b, avgLineLength, func(r io.Reader) Scanner {
		return bufio.NewScanner(r)
	})
}

func BenchmarkScannerAverageGobls(b *testing.B) {
	benchmarkScanner(b, avgLineLength, func(r io.Reader) Scanner {
		return NewScanner(r)
	})
}

func BenchmarkScannerShortBufio(b *testing.B) {
	benchmarkScanner(b, shortLineLength, func(r io.Reader) Scanner {
		return bufio.NewScanner(r)
	})
}

func BenchmarkScannerShortGobls(b *testing.B) {
	benchmarkScanner(b, shortLineLength, func(r io.Reader) Scanner {
		return NewScanner(r)
	})
}

func BenchmarkScannerLongBufio(b *testing.B) {
	benchmarkScanner(b, longLineLength, func(r io.Reader) Scanner {
		return bufio.NewScanner(r)
	})
}

func BenchmarkScannerLongGobls(b *testing.B) {
	benchmarkScanner(b, longLineLength, func(r io.Reader) Scanner {
		return NewScanner(r)
	})
}

func BenchmarkScannerVeryLongBufio(b *testing.B) {
	benchmarkScanner(b, veryLongLineLength, func(r io.Reader) Scanner {
		return bufio.NewScanner(r)
	})
}

func BenchmarkScannerVeryLongGobls(b *testing.B) {
	benchmarkScanner(b, veryLongLineLength, func(r io.Reader) Scanner {
		return NewScanner(r)
	})
}
