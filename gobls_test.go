package gobls

import (
	"bufio"
	"bytes"
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

func ensureDone(tb testing.TB, s Scanner) {
	tb.Helper()

	// Scan and check results.
	if got, want := s.Scan(), false; got != want {
		tb.Errorf("GOT: %#v; WANT: %#v", got, want)
	}
	if got, want := s.Text(), ""; got != want {
		tb.Errorf("GOT: %#v; WANT: %#v", got, want)
	}
	if got, want := s.Err(), error(nil); got != want {
		tb.Errorf("GOT: %#v; WANT: %#v", got, want)
	}

	//  Do it again to ensure idempotent.
	if got, want := s.Scan(), false; got != want {
		tb.Errorf("GOT: %#v; WANT: %#v", got, want)
	}
	if got, want := s.Text(), ""; got != want {
		tb.Errorf("GOT: %#v; WANT: %#v", got, want)
	}
	if got, want := s.Err(), error(nil); got != want {
		tb.Errorf("GOT: %#v; WANT: %#v", got, want)
	}
}

func ensureScan(tb testing.TB, s Scanner, v string) {
	tb.Helper()
	if got, want := s.Scan(), true; got != want {
		tb.Errorf("GOT: %#v; WANT: %#v", got, want)
	}
	if got, want := s.Text(), v; got != want {
		tb.Errorf("GOT: %#v; WANT: %#v", got, want)
	}
	if got, want := s.Err(), error(nil); got != want {
		tb.Errorf("GOT: %#v; WANT: %#v", got, want)
	}
}

func ensureSequence(tb testing.TB, s Scanner, seq []string) {
	tb.Helper()
	for _, want := range seq {
		ensureScan(tb, s, want)
	}
	ensureDone(tb, s)
}

func TestEmpty(t *testing.T) {
	corpus := ""
	ensureSequence(t, bufio.NewScanner(bytes.NewBufferString(corpus)), nil)
	ensureSequence(t, newBufferScanner([]byte(corpus)), nil)
}

func TestSequencesThroughEntireBuffer(t *testing.T) {
	corpus := "flubber\nblubber\nfoo"
	want := []string{"flubber", "blubber", "foo"}

	ensureSequence(t, bufio.NewScanner(bytes.NewBufferString(corpus)), want)
	ensureSequence(t, NewScanner(bytes.NewBufferString(corpus)), want)
	ensureSequence(t, newBufferScanner([]byte(corpus)), want)
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
		if got, want := s.Err(), error(nil); got != want {
			t.Errorf("GOT: %#v; WANT: %#v", got, want)
		}
		if got, want := len(lines), 1; got != want {
			t.Fatalf("GOT: %#v; WANT: %#v", got, want)
		}
		if got, want := lines[0], line; got != want {
			t.Errorf("GOT: %#v; WANT: %#v", got, want)
		}
	}

	test(bufio.NewScanner(bytes.NewReader(buf)))
	test(NewScanner(bytes.NewReader(buf)))
	test(newBufferScanner([]byte(buf)))
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
		if got, want := s.Err(), error(nil); got != want {
			t.Errorf("GOT: %#v; WANT: %#v", got, want)
		}
		if got, want := len(lines), 1; got != want {
			t.Fatalf("GOT: %#v; WANT: %#v", got, want)
		}
		if got, want := lines[0], line; got != want {
			t.Errorf("GOT: %#v; WANT: %#v", got, want)
		}
	}

	// test(bufio.NewScanner(bytes.NewReader(buf))) // bufio will return err = token too long
	test(NewScanner(bytes.NewReader(buf)))
	test(newBufferScanner([]byte(buf)))
}

func benchmarkScanner(b *testing.B, lineLength int, makeScanner func(buf []byte) Scanner) {
	b.Helper()

	wanted := makeBuffer(1, lineLength)
	wanted = wanted[:len(wanted)-2] // trim CRLF

	// NOTE: make buffer with line count set to b.N
	s := makeScanner(makeBuffer(b.N, lineLength))

	var line []byte
	var count int

	b.ResetTimer()

	for s.Scan() {
		line = s.Bytes()
		count++
	}

	if got, want := s.Err(), error(nil); got != want {
		b.Errorf("GOT: %#v; WANT: %#v", got, want)
	}
	// NOTE: ensure proper number of lines scanned
	if got, want := count, b.N; got != want {
		b.Errorf("GOT: %#v; WANT: %#v", got, want)
	}
	// NOTE: test line contents to prevent compiler optimization from eliding call to s.Bytes()
	if !bytes.Equal(line, wanted) {
		b.Fatalf("GOT: %#v; WANT: %#v", line, wanted)
	}
}

func BenchmarkScannerAverage(b *testing.B) {
	b.Run("bufio", func(b *testing.B) {
		benchmarkScanner(b, avgLineLength, func(buf []byte) Scanner {
			return bufio.NewScanner(bytes.NewReader(buf))
		})
	})
	b.Run("reader", func(b *testing.B) {
		benchmarkScanner(b, avgLineLength, func(buf []byte) Scanner {
			return NewScanner(bytes.NewReader(buf))
		})
	})
	b.Run("reader", func(b *testing.B) {
		benchmarkScanner(b, avgLineLength, func(buf []byte) Scanner {
			return newBufferScanner(buf)
		})
	})
}

func BenchmarkScannerShort(b *testing.B) {
	b.Run("bufio", func(b *testing.B) {
		benchmarkScanner(b, shortLineLength, func(buf []byte) Scanner {
			return bufio.NewScanner(bytes.NewReader(buf))
		})
	})
	b.Run("reader", func(b *testing.B) {
		benchmarkScanner(b, shortLineLength, func(buf []byte) Scanner {
			return NewScanner(bytes.NewReader(buf))
		})
	})
}

func BenchmarkScannerLong(b *testing.B) {
	b.Run("bufio", func(b *testing.B) {
		benchmarkScanner(b, longLineLength, func(buf []byte) Scanner {
			return bufio.NewScanner(bytes.NewReader(buf))
		})
	})
	b.Run("reader", func(b *testing.B) {
		benchmarkScanner(b, longLineLength, func(buf []byte) Scanner {
			return NewScanner(bytes.NewReader(buf))
		})
	})
}

func BenchmarkScannerVeryLong(b *testing.B) {
	b.Run("bufio", func(b *testing.B) {
		benchmarkScanner(b, veryLongLineLength, func(buf []byte) Scanner {
			return bufio.NewScanner(bytes.NewReader(buf))
		})
	})
	b.Run("reader", func(b *testing.B) {
		benchmarkScanner(b, veryLongLineLength, func(buf []byte) Scanner {
			return NewScanner(bytes.NewReader(buf))
		})
	})
}
