package llmclient

import (
	"time"
)

// defaultAfter wraps time.After so it can be replaced in tests.
var defaultAfter = func(d time.Duration) <-chan time.Time {
	return time.After(d)
}
