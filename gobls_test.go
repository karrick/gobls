package gobls

import (
	"bytes"
	"testing"
)

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

func TestFoo(t *testing.T) {
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
