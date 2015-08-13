# gobls

Gobls is a buffered line scanner for Go.

[![GoDoc](https://godoc.org/github.com/karrick/gobls?status.svg)](https://godoc.org/github.com/karrick/gobls)

## Description

Similar to `bufio.Scanner`, but wraps `bufio.ReadLine` so lines of
arbitrary length can be scanned. Uses a hybrid approach so that in
most cases, when lines are not unusually long, the fast code path is
taken. When lines are unusually long, it uses a per-Scanner
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

Gobls is approximately one quarter the speed of the standard
library. While the `Scan()` method for `bufio.Scanner` returns in
fewer than 30 nanoseconds on my test system for most line lengths, it
takes gobls around 100 nanoseconds on the same system under similar
load. For this reason, I recommend using `bufio.Scanner` for programs
unless a specific program must be able to parse lines that exceed a
very large constant, `bufio.MaxScanTokenSize`. In this case, the
additional delay due to extremely long lines may be an acceptible
tradeoff compared to the errors that would be returned by
`bufio.Scanner`.
