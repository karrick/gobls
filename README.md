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

## Performance

On my test system, gobls scanner takes from 2% to nearly 40% longer
than bufio scanner, depending on the length of the lines to be
scanned.  The 40% longer times were only observed when line lengths
were `bufio.MaxScanTokenSize` bytes long.  Usually the performance
penalty is 2% to 15% of bufio measurements.

Run `go test -bench=.` on your system for comparison.  I'm sure the
testing method could be improved.  Suggestions are welcomed.

I recommend using standard library's bufio scanner for programs unless
a specific program must be able to parse lines that exceed a very
large constant, `bufio.MaxScanTokenSize`. In this case, the additional
delay due to extremely long lines may be an acceptible tradeoff
compared to the errors that would be returned by `bufio.Scanner`.
