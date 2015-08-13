# gobls

Gobls is a buffered line scanner for Go.

[![GoDoc](https://godoc.org/github.com/karrick/gobls?status.svg)](https://godoc.org/github.com/karrick/gobls)

## Description

Similar to `bufio.Scanner`, but wraps `bufio.ReadLine` so lines of
arbitrary length can be scanned. Uses a hybrid approach so that in
most cases, when lines are not unusually long, the fast code path is
taken. When lines are unusually long, it uses `bytes.Buffer` to
reassemble the fragments into a single slice of bytes.

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

Gobls is within a few microseconds of `bufio.Scanner` for most line
lengths, imposing a slight performance penalty on the caller. For this
reason, you really should just use `bufio.Scanner` unless your program
must be able to parse lines that exceed a very large constant,
`bufio.MaxScanTokenSize`.

For lines with fewer than 4096 bytes, however, gobls is quite
performant, using the `ReadLine()` method of `bufio.Reader` to do most
of the work.

Gobls uses a pool of large pre-allocated `bytes.Buffer` objects to
quickly read additional data when line lengths exceed 4096 bytes.

If you know your line lengths are all much larger than 4096 bytes,
consider using `gobls.NewScannerSize(r io.Reader, size int)` to
specify a buffer size. Using this method, gobls is just as fast as
`bufio.Scanner`, and in some cases more fast.

```Go
    var lines, characters int
    ls := gobls.NewScannerSize(os.Stdin, 32768)
    for ls.Scan() {
        lines++
        characters += len(ls.Bytes())
    }
    if err:= ls.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "cannot scan:", err)
    }
    fmt.Println("Counted",lines,"lines and",characters,"characters.")
```
