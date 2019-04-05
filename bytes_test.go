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

func TestResultsSplit(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		results := &Results{buf: []byte("\n")}
		ensureStringSlicesMatch(t, results.Split(), nil)
	})
	t.Run("single", func(t *testing.T) {
		results := &Results{buf: []byte("one\n")}
		ensureStringSlicesMatch(t, results.Split(), []string{"one"})
	})
	t.Run("double", func(t *testing.T) {
		results := &Results{buf: []byte("one\ntwo\n")}
		ensureStringSlicesMatch(t, results.Split(), []string{"one", "two"})
	})
	t.Run("triple", func(t *testing.T) {
		results := &Results{buf: []byte("one\ntwo\nthree\n")}
		ensureStringSlicesMatch(t, results.Split(), []string{"one", "two", "three"})
	})
}

////////////////////////////////////////
// Scan
//

func scanit(results *Results) []string {
	var slice []string
	for results.Scan() {
		slice = append(slice, results.Text())
	}
	return slice
}

func ExampleResultsScan() {
	results, err := newResultsFromReader(bytes.NewReader([]byte("0\n1\n2\n3\n4\n5\n6\n7\n8\n9\n")))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var i int
	for results.Scan() {
		fmt.Printf("%s|", results.Text())
		if i++; i > 5 {
			break
		}
	}
	// Output: 0|1|2|3|4|5|
}

func TestResultsScan(t *testing.T) {
	t.Run("single newline", func(t *testing.T) {
		results := &Results{buf: []byte("\n")}
		ensureStringSlicesMatch(t, scanit(results), nil)
	})
	t.Run("single response", func(t *testing.T) {
		results := &Results{buf: []byte("r0\n")}
		ensureStringSlicesMatch(t, scanit(results), []string{"r0"})
	})
	t.Run("double response", func(t *testing.T) {
		results := &Results{buf: []byte("r0\nr1\n")}
		ensureStringSlicesMatch(t, scanit(results), []string{"r0", "r1"})
	})
	t.Run("triple response", func(t *testing.T) {
		results := &Results{buf: []byte("r0\nr1\nr2\n")}
		ensureStringSlicesMatch(t, scanit(results), []string{"r0", "r1", "r2"})
	})
}

////////////////////////////////////////
// benchmarks

const benchmarkCount = 10000000

var benchmarkLongResults []byte

func setupLongResults() {
	if benchmarkLongResults == nil {
		for i := int64(0); i < benchmarkCount; i++ {
			benchmarkLongResults = append(strconv.AppendInt(benchmarkLongResults, i, 16), '\n')
		}
	}
}

func testScanLarge(tb testing.TB) {
	r := &Results{buf: benchmarkLongResults}
	slice := scanit(r)
	if got, want := len(slice), benchmarkCount; got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := slice[0], "0"; got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := slice[len(slice)-1], strconv.FormatInt(benchmarkCount-1, 16); got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
}

func testSplitLarge(tb testing.TB) {
	r := &Results{buf: benchmarkLongResults}
	slice := r.Split()
	if got, want := len(slice), benchmarkCount; got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := slice[0], "0"; got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
	if got, want := slice[len(slice)-1], strconv.FormatInt(benchmarkCount-1, 16); got != want {
		tb.Fatalf("GOT: %v; WANT: %v", got, want)
	}
}

func BenchmarkResultsScanLarge(b *testing.B) {
	setupLongResults()
	b.ResetTimer()

	if benchmarkLongResults != nil {
		for i := 0; i < b.N; i++ {
			testScanLarge(b)
		}
	}
}

func BenchmarkResultsSplitLarge(b *testing.B) {
	setupLongResults()
	b.ResetTimer()

	if benchmarkLongResults != nil {
		for i := 0; i < b.N; i++ {
			testSplitLarge(b)
		}
	}
}
