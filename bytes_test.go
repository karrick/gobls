package gobls

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"testing"
)

func ensureStringSlicesMatch(tb testing.TB, actual, expected []string) {
	tb.Helper()
	if got, want := len(actual), len(expected); got != want {
		tb.Errorf("GOT: %v; WANT: %v", got, want)
	}
	la := len(actual)
	le := len(expected)
	for i := 0; i < la || i < le; i++ {
		if i < la {
			if i < le {
				if got, want := actual[i], expected[i]; got != want {
					tb.Errorf("GOT: %q; WANT: %q", got, want)
				}
			} else {
				tb.Errorf("GOT: %q (extra)", actual[i])
			}
		} else if i < le {
			tb.Errorf("WANT: %q (missing)", expected[i])
		}
	}
}

func TestNewResponseSplit(t *testing.T) {
	run := func(tb testing.TB, input string, expected []string) {
		tb.Helper()
		r, err := newResponseFromReader(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatal(err)
		}
		ensureStringSlicesMatch(tb, r.Split(), expected)
	}
	t.Run("empty", func(t *testing.T) {
		run(t, "", nil)
		run(t, "\n", nil)
	})
	t.Run("single", func(t *testing.T) {
		run(t, "one", []string{"one"})
		run(t, "one\n", []string{"one"})
	})
	t.Run("double", func(t *testing.T) {
		run(t, "one\ntwo", []string{"one", "two"})
		run(t, "one\ntwo\n", []string{"one", "two"})
	})
	t.Run("with empty string", func(t *testing.T) {
		run(t, "one\n\nthree", []string{"one", "", "three"})
		run(t, "one\n\nthree\n", []string{"one", "", "three"})
	})
}

////////////////////////////////////////
// Scan
////////////////////////////////////////

func ExampleResponseScan() {
	responses, err := newResponseFromReader(bytes.NewReader([]byte("0\n1\n2\n3\n4\n5\n6\n7\n8\n9\n")))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var i int
	for responses.Scan() {
		fmt.Printf("%s|", responses.Text())
		if i++; i > 5 {
			break
		}
	}
	// Output: 0|1|2|3|4|5|
}

type scannerI interface {
     Scan() bool
     Text() string
}

func scanit(r scannerI) []string {
	var values []string
	for r.Scan() {
		values = append(values, r.Text())
	}
	return values
}

func TestNewResponseScan(t *testing.T) {
	run := func(tb testing.TB, input string, expected []string) {
		tb.Helper()
		r, err := newResponseFromReader(bytes.NewReader([]byte(input)))
		if err != nil {
			t.Fatal(err)
		}
		ensureStringSlicesMatch(tb, scanit(r), expected)
	}
	t.Run("empty", func(t *testing.T) {
		run(t, "", nil)
		run(t, "\n", nil)
	})
	t.Run("single", func(t *testing.T) {
		run(t, "one", []string{"one"})
		run(t, "one\n", []string{"one"})
	})
	t.Run("double", func(t *testing.T) {
		run(t, "one\ntwo", []string{"one", "two"})
		run(t, "one\ntwo\n", []string{"one", "two"})
	})
	t.Run("with empty string", func(t *testing.T) {
		run(t, "one\n\nthree", []string{"one", "", "three"})
		run(t, "one\n\nthree\n", []string{"one", "", "three"})
	})
}

////////////////////////////////////////
// benchmarks

const benchmarkCount = 1000000

var benchmarkLongResponse []byte
var benchmarkLongValues []string

func setupLongResponse() {
	if benchmarkLongResponse == nil {
        benchmarkLongValues = make([]string, benchmarkCount)
		for i := int64(0); i < benchmarkCount; i++ {
            v := strconv.FormatInt(i, 16)
            benchmarkLongValues[i] = v
		    benchmarkLongResponse = append(benchmarkLongResponse, v + "\n"...)
		}
        // Trim final newline from byte slice.
        benchmarkLongResponse = benchmarkLongResponse[:len(benchmarkLongResponse)-1]
	}
}

func testScanLarge(tb testing.TB) {
	r := &response{buf: benchmarkLongResponse}
	values := scanit(r)
	if got, want := len(values), benchmarkCount; got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := values[0], "0"; got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := values[len(values)-1], strconv.FormatInt(benchmarkCount-1, 16); got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
}

func testScanReaderLarge(tb testing.TB) {
    s := NewScanner(bytes.NewReader(benchmarkLongResponse))
	values := scanit(s)
	if got, want := len(values), benchmarkCount; got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := values[0], "0"; got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := values[len(values)-1], strconv.FormatInt(benchmarkCount-1, 16); got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
}

func testSplitLarge(tb testing.TB) {
	r := &response{buf: benchmarkLongResponse}
	values := r.Split()
	if got, want := len(values), benchmarkCount; got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := values[0], "0"; got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := values[len(values)-1], strconv.FormatInt(benchmarkCount-1, 16); got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
}

func BenchmarkResponseScanReaderLarge(b *testing.B) {
	setupLongResponse()
	b.ResetTimer()

	if benchmarkLongResponse != nil {
		for i := 0; i < b.N; i++ {
			testScanReaderLarge(b)
		}
	}
}

func BenchmarkResponseScanLarge(b *testing.B) {
	setupLongResponse()
	b.ResetTimer()

	if benchmarkLongResponse != nil {
		for i := 0; i < b.N; i++ {
			testScanLarge(b)
		}
	}
}

func BenchmarkResponseSplitLarge(b *testing.B) {
	setupLongResponse()
	b.ResetTimer()

	if benchmarkLongResponse != nil {
		for i := 0; i < b.N; i++ {
			testSplitLarge(b)
		}
	}
}
