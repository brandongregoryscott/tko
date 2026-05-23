package grid

import (
	"io"
	"log"
	"os"
)

// debug is the package-level debug logger. It writes to io.Discard by
// default, so debug output is suppressed unless explicitly enabled via
// SetDebug. This keeps the terminal clean when no grid is connected.
var debug = log.New(io.Discard, "[grid] ", log.Ltime)

// SetDebug enables or disables debug logging for the grid package.
// When enabled, debug logs are written to stderr.
func SetDebug(enabled bool) {
	if enabled {
		debug.SetOutput(os.Stderr)
	} else {
		debug.SetOutput(io.Discard)
	}
}
