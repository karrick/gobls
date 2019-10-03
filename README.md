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

The `BufferScanner` is faster than `bufio.Scanner` for all
benchmarks. However, on my test system, the regular `Scanner` takes
from 2% to nearly 40% longer than bufio scanner, depending on the
length of the lines to be scanned.  The 40% longer times were only
observed when line lengths were `bufio.MaxScanTokenSize` bytes long.
Usually the performance penalty is 2% to 15% of bufio measurements.

Run `go test -bench=. -benchmem` on your system for comparison.  I'm
sure the testing method could be improved.  Suggestions are welcomed.

For circumstances where there is no concern about enumerating lines
whose lengths are longer than the max token length from `bufio`, then
I recommend using the standard library.

On the other hand, if you already have a slice of bytes, library is
much more performant than the equivalent
`bufio.NewScanner(bytes.NewReader(buf))`.

```
$ go test -bench=. -benchmem
goos: linux
goarch: amd64
pkg: github.com/karrick/gobls
BenchmarkScanner/0064/bufio-8               30000000   43.7  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/0064/reader-8              20000000   59.2  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/0064/buffer-8              50000000   33.7  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/0128/bufio-8               30000000   54.5  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/0128/reader-8              20000000   70.5  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/0128/buffer-8              30000000   38.9  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/0256/bufio-8               20000000   79.8  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/0256/reader-8              20000000   94.9  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/0256/buffer-8              30000000   50.2  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/0512/bufio-8               10000000    123  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/0512/reader-8              10000000    144  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/0512/buffer-8              20000000   79.0  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/1024/bufio-8               10000000    210  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/1024/reader-8              10000000    227  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/1024/buffer-8              10000000    119  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/2048/bufio-8                5000000    382  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/2048/reader-8               3000000    413  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/2048/buffer-8               5000000    272  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/4096/bufio-8                2000000    701  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/4096/reader-8               2000000    733  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/4096/buffer-8               3000000    517  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/excessively_long/bufio-8     200000  11681  ns/op  0  B/op  0  allocs/op
BenchmarkScanner/excessively_long/reader-8    100000  14464  ns/op  2  B/op  0  allocs/op
BenchmarkScanner/excessively_long/buffer-8    200000   8688  ns/op  0  B/op  0  allocs/op
PASS
ok  	github.com/karrick/gobls	256.191s
```
