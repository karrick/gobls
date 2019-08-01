// +build gobls_debug

package gobls

import (
	"fmt"
	"os"
)

// debug formats and prints arguments to stderr for development builds
func debug(f string, a ...interface{}) {
	os.Stderr.Write([]byte("gobls: " + fmt.Sprintf(f, a...)))
}
