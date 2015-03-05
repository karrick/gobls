# gobls

Gobls is a buffered line scanner for Go.

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
    if ls.Err() != nil {
        fmt.Fprintln(os.Stderr, "cannot scan:", ls.Err())
    }
    fmt.Println("Counted",lines,"and",characters,"characters.")
```
