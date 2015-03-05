package gobls

import (
	"bytes"
	"fmt"
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
	for s.Scan() {
		fmt.Println(s.String())
	}
	if s.Err() != nil {
		t.Errorf("Actual: %#v; Expected: %#v", s.Err(), nil)
	}
}
