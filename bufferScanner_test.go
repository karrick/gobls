package gobls

import "testing"

func TestBufferScanner(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		s := NewBufferScanner(nil)
		ensureDone(t, s)
		ensureDone(t, s)
	})

	t.Run("single byte", func(t *testing.T) {
		t.Run("newline", func(t *testing.T) {
			s := NewBufferScanner([]byte{'\n'})
			ensureScan(t, s, "")
			ensureDone(t, s)
		})
		t.Run("carriage return", func(t *testing.T) {
			s := NewBufferScanner([]byte{'\r'})
			ensureScan(t, s, "")
			ensureDone(t, s)
		})
		t.Run("other", func(t *testing.T) {
			s := NewBufferScanner([]byte{'a'})
			ensureScan(t, s, "a")
			ensureDone(t, s)
		})
	})

	t.Run("single line", func(t *testing.T) {
		t.Run("with newline", func(t *testing.T) {
			s := NewBufferScanner([]byte("line1\n"))
			ensureScan(t, s, "line1")
			ensureDone(t, s)
		})
		t.Run("with carriage return", func(t *testing.T) {
			s := NewBufferScanner([]byte("line1\r"))
			ensureScan(t, s, "line1")
			ensureDone(t, s)
		})
		t.Run("with both", func(t *testing.T) {
			s := NewBufferScanner([]byte("line1\r\n"))
			ensureScan(t, s, "line1")
			ensureDone(t, s)
		})
		t.Run("with neither", func(t *testing.T) {
			s := NewBufferScanner([]byte("line1"))
			ensureScan(t, s, "line1")
			ensureDone(t, s)
		})
	})

	t.Run("double line", func(t *testing.T) {
		t.Run("with trailing newline", func(t *testing.T) {
			s := NewBufferScanner([]byte("line1\nline2\n"))
			ensureScan(t, s, "line1")
			ensureScan(t, s, "line2")
			ensureDone(t, s)
		})
		t.Run("with trailing carriage return", func(t *testing.T) {
			// NOTE: Because carriage returns are ignored rather than marking
			// the end of a line, this source buffer returns a single line.
			s := NewBufferScanner([]byte("line1\rline2\r"))
			ensureScan(t, s, "line1\rline2")
			ensureDone(t, s)
		})
		t.Run("with trailing both", func(t *testing.T) {
			s := NewBufferScanner([]byte("line1\r\nline2\r\n"))
			ensureScan(t, s, "line1")
			ensureScan(t, s, "line2")
			ensureDone(t, s)
		})
	})
}
