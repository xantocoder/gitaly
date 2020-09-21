package safe

import (
	"sync"
	"time"
)

// Threshold returns a function each call of which returns a flag if
// passed in threshold is reached in window time duration.
// If threshold=3 and window=1s and function called 3 times within 1s
// time interval then the last call will return true. This is a signal
// that the threshold of invocation was reached in a configured time window.
func Threshold(threshold int, window time.Duration) func() bool {
	var triggeredAt []time.Time
	var mtx sync.Mutex

	return func() bool {
		mtx.Lock()
		defer mtx.Unlock()

		now := time.Now()
		triggeredAt = append(triggeredAt, now)
		if len(triggeredAt) > threshold {
			triggeredAt = triggeredAt[len(triggeredAt)-threshold:]
		}

		inRange := 1
		for i := len(triggeredAt) - 2; i >= 0; i-- {
			if now.Sub(triggeredAt[i]) <= window {
				inRange++
			}
		}

		return inRange >= threshold
	}
}
