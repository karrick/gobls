# gobls

Gobls is a buffered line scanner for Go.

[![GoDoc](https://godoc.org/github.com/karrick/gobls?status.svg)](https://godoc.org/github.com/karrick/gobls)

## Description

Similar to `bufio.Scanner`, but wraps `bufio.Reader.ReadLine` so lines
of arbitrary length can be scanned.  It uses a hybrid approach so that
in most cases, when lines are not unusually long, the fast code path
is taken.  When lines are unusually long, it uses the per-scanner
pre-allocated byte slice to reassemble the fragments into a single
slice of bytes.

## Example

### Enumerating lines from an io.Reader (drop in replacement for bufio.Scanner)

When you have an io.Reader that you want to enumerate, normally you
wrap it in `bufio.Scanner`. This library is a drop in replacement for
this particular circumstance, and you can change from
`bufio.NewScanner(r)` to `gobls.NewScanner(r)`, and no longer have to
worry about token too long errors.

```Go
    var lines, characters int
    ls := gobls.NewScanner(os.Stdin)
    for ls.Scan() {
        lines++
        characters += len(ls.Bytes())
    }
    if err:= ls.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "cannot scan:", err)
    }
    fmt.Println("Counted",lines,"lines and",characters,"characters.")
```

### Enumerating lines from []byte

If you already have a slice of bytes that you want to enumerate lines
for, it is much more performant to wrap that byte slice with
`gobls.NewBufferScanner(buf)` than to wrap the slice in a io.Reader
and call either the above or `bufio.NewScanner`.

```Go
    var lines, characters int
    ls := gobls.NewBufferScanner(buf)
    for ls.Scan() {
        lines++
        characters += len(ls.Bytes())
    }
    if err:= ls.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "cannot scan:", err)
    }
    fmt.Println("Counted",lines,"lines and",characters,"characters.")
```

## Performance

On my test system, gobls scanner takes from 2% to nearly 40% longer
than bufio scanner, depending on the length of the lines to be
scanned.  The 40% longer times were only observed when line lengths
were `bufio.MaxScanTokenSize` bytes long.  Usually the performance
penalty is 2% to 15% of bufio measurements.

Run `go test -bench=.` on your system for comparison.  I'm sure the
testing method could be improved.  Suggestions are welcomed.

For circumstances where there is no concern about enumerating lines
whose lengths are longer than the max token length from `bufio`, then
I recommend using the standard library. However if you already have a
slice of bytes, this library is much more performant than the
equivalent `bufio.NewScanner(bytes.NewReader(buf))`.

```
$ go test -bench=. -benchmem
goos: linux
goarch: amd64
pkg: github.com/karrick/gobls
BenchmarkScannerAverage/bufio-12    10000000    198  ns/op  0  B/op  0  allocs/op
BenchmarkScannerAverage/reader-12   10000000    199  ns/op  0  B/op  0  allocs/op
BenchmarkScannerAverage/buffer-12   10000000    122  ns/op  0  B/op  0  allocs/op
BenchmarkScannerShort/bufio-12      30000000   45.4  ns/op  0  B/op  0  allocs/op
BenchmarkScannerShort/reader-12     30000000   56.5  ns/op  0  B/op  0  allocs/op
BenchmarkScannerShort/buffer-12     50000000   37.7  ns/op  0  B/op  0  allocs/op
BenchmarkScannerLong/bufio-12        2000000    614  ns/op  0  B/op  0  allocs/op
BenchmarkScannerLong/reader-12       2000000    628  ns/op  0  B/op  0  allocs/op
BenchmarkScannerLong/buffer-12       5000000    379  ns/op  0  B/op  0  allocs/op
BenchmarkScannerVeryLong/bufio-12     200000   9616  ns/op  0  B/op  0  allocs/op
BenchmarkScannerVeryLong/reader-12    100000  13177  ns/op  2  B/op  0  allocs/op
BenchmarkScannerVeryLong/buffer-12    200000   6163  ns/op  0  B/op  0  allocs/op
PASS
ok      github.com/karrick/gobls        159.336s
```
