package gobls

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func benchmarkScanner(b *testing.B, lineLength int, makeScanner func(buf []byte) Scanner) {
	b.Helper()

	const lineCount = 100
	wanted := makeBuffer(1, lineLength)
	wanted = wanted[:len(wanted)-2] // trim final CRLF
	buf := makeBuffer(lineCount, lineLength)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var count int
		s := makeScanner(buf)
		for s.Scan() {
			if got := s.Bytes(); !bytes.Equal(got, wanted) {
				b.Errorf("GOT: %#v; WANT: %#v", got, wanted)
			}
			count++
		}
		if got, want := s.Err(), error(nil); got != want {
			b.Fatalf("GOT: %#v; WANT: %#v", got, want)
		}
		if got, want := count, lineCount; got != want {
			b.Fatalf("GOT: %#v; WANT: %#v", got, want)
		}
	}
}

func BenchmarkScanner(b *testing.B) {
	for _, i := range []uint{6, 7, 8, 9, 10, 11, 12} {
		lineLength := 1 << i

		b.Run(fmt.Sprintf("%04d", lineLength), func(b *testing.B) {
			b.Run("bufio", func(b *testing.B) {
				benchmarkScanner(b, lineLength, func(buf []byte) Scanner {
					return bufio.NewScanner(bytes.NewReader(buf))
				})
			})
			b.Run("scanner", func(b *testing.B) {
				benchmarkScanner(b, lineLength, func(buf []byte) Scanner {
					return NewScanner(bytes.NewReader(buf))
				})
			})
			b.Run("buffer", func(b *testing.B) {
				benchmarkScanner(b, lineLength, func(buf []byte) Scanner {
					return NewBufferScanner(buf)
				})
			})
		})
	}

	b.Run("excessively long", func(b *testing.B) {
		const lineLength = excessivelyLongLineLength
		b.Run("bufio", func(b *testing.B) {
			benchmarkScanner(b, lineLength, func(buf []byte) Scanner {
				return bufio.NewScanner(bytes.NewReader(buf))
			})
		})
		b.Run("scanner", func(b *testing.B) {
			benchmarkScanner(b, lineLength, func(buf []byte) Scanner {
				return NewScanner(bytes.NewReader(buf))
			})
		})
		b.Run("buffer", func(b *testing.B) {
			benchmarkScanner(b, lineLength, func(buf []byte) Scanner {
				return NewBufferScanner(buf)
			})
		})
	})
}
