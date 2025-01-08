package gobls

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Use a time format that is both RFC-3339 and ISO-8601 compliant:
// %Y-%M-%DT%h:%m:%sZ
const time_format = "2006-01-02T15:04:05Z07:00"

const fileMode = 0640 // rw-r-----

var isDebug = false

// debug either ignores its arguments or uses them to print debugging
// information to standard error.
var debug func(string, ...interface{}) = func(string, ...interface{}) {}

func init() {
	var err error
	var fh io.WriteCloser

	_ = isDebug

	if location := os.Getenv("GOBLS_DEBUG"); location != "" {
		switch strings.ToLower(location) {
		case "stderr":
			fh = os.Stderr
		case "stdout":
			fh = os.Stdout
		default:
			fh, err = os.OpenFile(location, os.O_WRONLY|os.O_CREATE|os.O_APPEND, fileMode)
			if err != nil {
				now := time.Now().Format(time_format)
				_, _ = fmt.Fprintf(os.Stderr, "%s GOBLS_DEBUG: cannot open file: %q: %s\n", now, location, err)
				fh = os.Stderr
			}
		}

		debug = func(f string, a ...interface{}) {
			now := time.Now().Format(time_format)
			content := fmt.Sprintf(f, a...)
			if f != "" && f[len(f)-1] == '\n' {
				_, _ = fmt.Fprintf(fh, "%s GOBLS_DEBUG: %s", now, content)
			} else {
				_, _ = fmt.Fprintf(fh, "%s GOBLS_DEBUG: %s\n", now, content)
			}
		}

		isDebug = true
	}
}
