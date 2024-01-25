package gobls

import (
	"bufio"
	"bytes"
	"testing"
)

const excessivelyLongLineLength = bufio.MaxScanTokenSize - 2

func makeBufioScanner(buf []byte) Scanner {
	return bufio.NewScanner(bytes.NewReader(buf))
}

func makeScanner(buf []byte) Scanner {
	return NewScanner(bytes.NewReader(buf))
}

func makeBufferScanner(buf []byte) Scanner {
	return NewBufferScanner(buf)
}

func makeBuffer(lineCount, lineLength int) []byte {
	buf := make([]byte, 0, lineCount*lineLength)
	for line := 0; line < lineCount; line++ {
		// Each line terminated by CRLF
		for i := 0; i < lineLength-2; i++ {
			buf = append(buf, 'a')
		}
		buf = append(buf, '\r', '\n')
	}
	return buf
}

func testSequenceBufio(t *testing.T, buf []byte, seq []string) {
	t.Helper()
	ensureSequence(t, "bufio", makeBufioScanner(buf), seq)
}

func testSequenceScanner(t *testing.T, buf []byte, seq []string) {
	t.Helper()
	ensureSequence(t, "scanner", makeScanner(buf), seq)
}

func testSequenceBuffer(t *testing.T, buf []byte, seq []string) {
	t.Helper()
	ensureSequence(t, "buffer", makeBufferScanner(buf), seq)
}

func TestEmpty(t *testing.T) {
	var corpus []byte
	var want []string

	testSequenceBufio(t, corpus, want)
	testSequenceScanner(t, corpus, want)
	testSequenceBuffer(t, corpus, want)
}

func TestSequencesThroughEntireBuffer(t *testing.T) {
	corpus := []byte("flubber\nblubber\nfoo")
	want := []string{"flubber", "blubber", "foo"}

	testSequenceBufio(t, corpus, want)
	testSequenceScanner(t, corpus, want)
	testSequenceBuffer(t, corpus, want)
}

func TestLongLinesRequireSingleInvocation(t *testing.T) {
	buf := makeBuffer(1, 4096)
	line := string(buf)
	line = line[:len(line)-2] // trim final CRLF
	want := []string{line}

	testSequenceBufio(t, buf, want)
	testSequenceScanner(t, buf, want)
	testSequenceBuffer(t, buf, want)
}

func TestHandlesLinesLongerThanBuffer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long buffer")
	}
	buf := makeBuffer(1, 8192)
	line := string(buf)
	line = line[:len(line)-2] // trim final CRLF
	want := []string{line}

	testSequenceBufio(t, buf, want)
	testSequenceScanner(t, buf, want)
	testSequenceBuffer(t, buf, want)
}

func TestExcessivelyLongLinesRequireSingleInvocation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long buffer")
	}
	buf := makeBuffer(1, bufio.MaxScanTokenSize+5)
	line := string(buf)
	line = line[:len(line)-2] // trim final CRLF
	want := []string{line}

	if false {
		// Test skipped because bufio will return err = token too long
		testSequenceBufio(t, buf, want)
	}
	testSequenceScanner(t, buf, want)
	testSequenceBuffer(t, buf, want)
}

func TestScannerEmpty(t *testing.T) {
	var buf []byte
	var want []string

	testSequenceBufio(t, buf, want)
	testSequenceScanner(t, buf, want)
	testSequenceBuffer(t, buf, want)
}

func TestScannerSingleByte(t *testing.T) {
	t.Run("newline", func(t *testing.T) {
		buf := []byte{'\n'}
		want := []string{""}

		testSequenceBufio(t, buf, want)
		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
	t.Run("carriage return", func(t *testing.T) {
		buf := []byte{'\r'}
		want := []string{""}

		testSequenceBufio(t, buf, want)
		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
	t.Run("other", func(t *testing.T) {
		buf := []byte{'a'}
		want := []string{"a"}

		testSequenceBufio(t, buf, want)
		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
}

func TestScannerSingleLine(t *testing.T) {
	t.Run("with newline", func(t *testing.T) {
		buf := []byte("line1\n")
		want := []string{"line1"}

		testSequenceBufio(t, buf, want)
		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
	t.Run("with carriage return", func(t *testing.T) {
		buf := []byte("line1\r")
		want := []string{"line1"}

		testSequenceBufio(t, buf, want)
		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
	t.Run("with both", func(t *testing.T) {
		buf := []byte("line1\r\n")
		want := []string{"line1"}

		testSequenceBufio(t, buf, want)
		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
	t.Run("with neither", func(t *testing.T) {
		buf := []byte("line1")
		want := []string{"line1"}

		testSequenceBufio(t, buf, want)
		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
}

func TestScannerDoubleLine(t *testing.T) {
	t.Run("with trailing newline", func(t *testing.T) {
		buf := []byte("line1\nline2\n")
		want := []string{"line1", "line2"}

		testSequenceBufio(t, buf, want)
		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
	t.Run("with trailing carriage return", func(t *testing.T) {
		// NOTE: Because carriage returns are ignored rather than marking
		// the end of a line, this source buffer returns a single line.
		buf := []byte("line1\rline2\r")
		want := []string{"line1\rline2"}

		testSequenceBufio(t, buf, want)
		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
	t.Run("with trailing both", func(t *testing.T) {
		buf := []byte("line1\r\nline2\r\n")
		want := []string{"line1", "line2"}

		testSequenceBufio(t, buf, want)
		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
}

func TestScannerLongLineWithoutEndline(t *testing.T) {
	t.Run("one long line", func(t *testing.T) {
		buf := makeBuffer(1, 1<<16)
		buf = buf[:len(buf)-2] // skip CLRF
		line := string(buf)
		want := []string{line}

		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
	t.Run("multiple long lines", func(t *testing.T) {
		buf := makeBuffer(3, 1<<16)
		buf = buf[:len(buf)-2] // skip CLRF
		line := string(buf[:(1<<16)-2])
		want := []string{line, line, line}

		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
	t.Run("long then short", func(t *testing.T) {
		buf := append(makeBuffer(1, 1<<16), makeBuffer(1, 1<<8)...)
		buf = buf[:len(buf)-2] // skip CLRF
		line1 := string(buf[:(1<<16)-2])
		line2 := string(buf[1<<16:])
		want := []string{line1, line2}

		testSequenceScanner(t, buf, want)
		testSequenceBuffer(t, buf, want)
	})
}
