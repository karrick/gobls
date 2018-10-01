package gobls

import (
	"bufio"
	"bytes"
	"io"
	"testing"
)

const (
	shortLineLength    = 64 - 2
	avgLineLength      = 1024 - 2
	longLineLength     = 4096 - 2
	veryLongLineLength = bufio.MaxScanTokenSize - 2
)

func makeBuffer(lineCount, lineLength int) []byte {
	buf := make([]byte, 0, lineCount*(lineLength+2))
	for line := 0; line < lineCount; line++ {
		for i := 0; i < lineLength; i++ {
			switch i % 10 {
			case 0:
				buf = append(buf, byte((i/10)%10+'0'))
			default:
				// buf = append(buf, '.')
				buf = append(buf, byte(i%10+'a'))
			}
		}
		buf = append(buf, '\r', '\n')
	}
	return buf
}

func TestCopy(t *testing.T) {
	buf := []byte("abcdefghijklmnopqrstuvwxyz")
	l := 20
	r := len(buf)
	// t.Logf("buf: %q", buf)
	nc := copy(buf, buf[l:r])
	if got, want := nc, 6; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	l = 0
	r = nc
	if got, want := l, 0; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := r, 6; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	buf = buf[l:r]
	if got, want := string(buf), "uvwxyz"; got != want {
		t.Errorf("GOT: %v; WANT: %v", got, want)
	}
	// t.Logf("buf: %q", buf)
}

func TestNoEOF(t *testing.T) {
	test := func(s Scanner) {
		if got, want := s.Scan(), false; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
		if got, want := s.Err(), error(nil); got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
	}

	corpus := ""
	test(bufio.NewScanner(bytes.NewBufferString(corpus)))
	test(NewScanner(bytes.NewBufferString(corpus)))
}

func TestSequencesThroughEntireBuffer(t *testing.T) {
	test := func(expected []string, s Scanner) {
		var gotLines []string
		for s.Scan() {
			gotLines = append(gotLines, s.Text())
		}
		if got, want := s.Err(), error(nil); got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
		if got, want := len(gotLines), len(expected); got != want {
			t.Fatalf("GOT: %v; WANT: %v", got, want)
		}
		for i := 0; i < len(expected); i++ {
			if got, want := gotLines[i], expected[i]; got != want {
				t.Errorf("GOT: %v; WANT: %v", got, want)
			}
		}
	}

	t.Run("LF", func(t *testing.T) {
		corpus := "flubber\nblubber\nfoo"
		expected := []string{"flubber", "blubber", "foo"}
		test(expected, bufio.NewScanner(bytes.NewBufferString(corpus)))
		test(expected, NewScanner(bytes.NewBufferString(corpus)))
	})

	t.Run("CRLF", func(t *testing.T) {
		corpus := "flubber\r\nblubber\r\nfoo"
		expected := []string{"flubber", "blubber", "foo"}
		test(expected, bufio.NewScanner(bytes.NewBufferString(corpus)))
		test(expected, NewScanner(bytes.NewBufferString(corpus)))
	})
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
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
		if got, want := len(lines), 1; got != want {
			t.Fatalf("GOT: %v; WANT: %v", got, want)
		}
		if got, want := lines[0], line; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
	}

	test(bufio.NewScanner(bytes.NewReader(buf)))
	test(NewScanner(bytes.NewReader(buf)))
}

func TestVeryLongLinesRequireSingleInvocation(t *testing.T) {
	buf := makeBuffer(1, bufio.MaxScanTokenSize+5)
	line := string(buf)
	line = line[:len(line)-2] // trim CRLF

	t.Run("bufio", func(t *testing.T) {
		// Expect bufio to yield error when scanning a line with more bytes than
		// bufio.MaxScanTokenSize.
		s := bufio.NewScanner(bytes.NewReader(buf))
		if got, want := s.Scan(), false; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
		if got, want := s.Err(), bufio.ErrTooLong; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
	})

	t.Run("gobls", func(t *testing.T) {
		lines := make([]string, 0, 1)
		s := NewScanner(bytes.NewReader(buf))
		for s.Scan() {
			lines = append(lines, s.Text())
		}
		if got, want := s.Err(), error(nil); got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
		if got, want := len(lines), 1; got != want {
			t.Fatalf("GOT: %v; WANT: %v", got, want)
		}
		if got, want := lines[0], line; got != want {
			t.Errorf("GOT: %v; WANT: %v", got, want)
		}
	})
}

func testScanner(tb testing.TB, lineCount, lineLength int, makeScanner func(io.Reader) Scanner) {
	tb.Helper()

	// Every line ought look like the following
	wanted := makeBuffer(1, lineLength)[:lineLength] // trim CRLF from tail of line

	s := makeScanner(bytes.NewReader(makeBuffer(lineCount, lineLength)))

	var count int
	var line []byte

	if b, ok := tb.(*testing.B); ok {
		b.ResetTimer()
		b.ReportAllocs()
	}

	for s.Scan() {
		count++
		line = s.Bytes()
	}

	if got, want := s.Err(), error(nil); got != want {
		tb.Errorf("GOT: %v; WANT: %v", got, want)
	}

	if got, want := count, lineCount; got != want {
		tb.Errorf("GOT: %v; GOT: %d", got, want)
	}

	if !bytes.Equal(line, wanted) {
		tb.Fatalf("GOT: %v; WANT: %v", string(line), string(wanted))
	}
}

func TestExtremelyLongLine(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	const lineCount = 100
	const lineSize = bufio.MaxScanTokenSize + 2

	t.Run("bufio", func(t *testing.T) {
		t.Skip("bufio.Scanner cannot process lines with more than bufio.MaxScanTokenSize bytes")
		testScanner(t, lineCount, lineSize, func(r io.Reader) Scanner {
			return bufio.NewScanner(r)
		})
	})

	t.Run("gobls", func(t *testing.T) {
		testScanner(t, lineCount, lineSize, func(r io.Reader) Scanner {
			return NewScanner(r)
		})
	})
}

func TestExtremelyManyLines(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	const lineCount = 50000000
	const lineSize = 50

	t.Run("bufio", func(t *testing.T) {
		testScanner(t, lineCount, lineSize, func(r io.Reader) Scanner {
			return bufio.NewScanner(r)
		})
	})

	t.Run("gobls", func(t *testing.T) {
		testScanner(t, lineCount, lineSize, func(r io.Reader) Scanner {
			return NewScanner(r)
		})
	})
}

func BenchmarkScannerShortBufio(b *testing.B) {
	testScanner(b, b.N, shortLineLength, func(r io.Reader) Scanner {
		return bufio.NewScanner(r)
	})
}

func BenchmarkScannerShortGobls(b *testing.B) {
	testScanner(b, b.N, shortLineLength, func(r io.Reader) Scanner {
		return NewScanner(r)
	})
}

func BenchmarkScannerAverageBufio(b *testing.B) {
	testScanner(b, b.N, avgLineLength, func(r io.Reader) Scanner {
		return bufio.NewScanner(r)
	})
}

func BenchmarkScannerAverageGobls(b *testing.B) {
	testScanner(b, b.N, avgLineLength, func(r io.Reader) Scanner {
		return NewScanner(r)
	})
}

func BenchmarkScannerLongBufio(b *testing.B) {
	testScanner(b, b.N, longLineLength, func(r io.Reader) Scanner {
		return bufio.NewScanner(r)
	})
}

func BenchmarkScannerLongGobls(b *testing.B) {
	testScanner(b, b.N, longLineLength, func(r io.Reader) Scanner {
		return NewScanner(r)
	})
}

func BenchmarkScannerVeryLongBufio(b *testing.B) {
	testScanner(b, b.N, veryLongLineLength, func(r io.Reader) Scanner {
		return bufio.NewScanner(r)
	})
}

func BenchmarkScannerVeryLongGobls(b *testing.B) {
	testScanner(b, b.N, veryLongLineLength, func(r io.Reader) Scanner {
		return NewScanner(r)
	})
}
