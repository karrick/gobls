package gobls

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"
)

func benchmarkScanner(b *testing.B, lineLength int, makeScanner func(buf []byte) Scanner) {
	b.Helper()

	wanted := makeBuffer(1, lineLength)
	wanted = wanted[:len(wanted)-2] // trim final CRLF

	// NOTE: make buffer with line count set to b.N
	s := makeScanner(makeBuffer(b.N, lineLength))

	var count int

	b.ResetTimer()

	for s.Scan() {
		if got := s.Bytes(); !bytes.Equal(got, wanted) {
			b.Errorf("GOT: %#v; WANT: %#v", got, wanted)
		}
		count++
	}

	if got, want := s.Err(), error(nil); got != want {
		b.Errorf("GOT: %#v; WANT: %#v", got, want)
	}
	if got, want := count, b.N; got != want {
		b.Errorf("GOT: %#v; WANT: %#v", got, want)
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
			b.Run("reader", func(b *testing.B) {
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
		b.Run("reader", func(b *testing.B) {
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
