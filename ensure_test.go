package gobls

import (
	"testing"
)

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

func ensureSequence(t *testing.T, name string, s Scanner, seq []string) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		t.Helper()
		for _, want := range seq {
			ensureScan(t, s, want)
		}
		ensureDone(t, s)
	})
}
