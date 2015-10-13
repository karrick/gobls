package gobls

// DefaultBufferSize specifies the initial bytes size each gobls scanner will allocate to be used
// for aggregation of line fragments.
const DefaultBufferSize = 16 * 1024

// Scanner provides an interface for reading newline-delimited lines of text.  It is similar to
// `bufio.Scanner`, but wraps the ReadLine method of `bufio.Reader` so lines of arbitrary length can
// be scanned.  Successive calls to the Scan method will step through the lines of a file, skipping
// the newline whitespace between lines.
//
// Scanning stops unrecoverably at EOF, or at the first I/O error.  Unlike `bufio.Scanner`, however,
// attempting to scan a line longer than `bufio.MaxScanTokenSize` will not result in an error, but
// will return the long line.
//
// It is not necessary to check for errors by calling the Err method until after scanning stops,
// when the Scan method returns false.
//
// Gobls Scanner ought behave exactly like `bufio.Scanner`.  All methods ought to have the exact
// same return values while stepping through the given `io.Reader`.
type Scanner interface {
	Bytes() []byte
	Err() error
	Scan() bool
	Text() string
}
