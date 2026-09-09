package coroutine

import (
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// DefaultFPS is the default frame rate used by the standalone clock.
const DefaultFPS = 30

var (
	clockFrame atomic.Int64
	clockNanos atomic.Int64
)

// Start resets the standalone clock. The argument is kept for compatibility
// with the original engine-backed implementation.
func Start(_ any) {
	clockFrame.Store(0)
	clockNanos.Store(0)
}

// Update advances the standalone clock by delta seconds and one frame.
func Update(delta float64, _ int) {
	if delta > 0 {
		clockNanos.Add(int64(delta * float64(time.Second)))
	}
	clockFrame.Add(1)
}

// Frame returns the current standalone scheduler frame.
func Frame() int64 {
	return clockFrame.Load()
}

// TimeSinceLevelLoad returns elapsed standalone level time in seconds.
func TimeSinceLevelLoad() float64 {
	return float64(clockNanos.Load()) / float64(time.Second)
}

// Sleep pauses the caller for seconds.
func Sleep(seconds float64) {
	if seconds > 0 {
		time.Sleep(time.Duration(seconds * float64(time.Second)))
	}
}

// currentGoroutineID returns the runtime ID of the calling goroutine. Go does
// not expose this as a public API, so this small parser is kept local to the
// standalone compatibility layer.
func currentGoroutineID() uint64 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	line := strings.TrimPrefix(string(buf[:n]), "goroutine ")
	if end := strings.IndexByte(line, ' '); end >= 0 {
		line = line[:end]
	}
	id, _ := strconv.ParseUint(line, 10, 64)
	return id
}
