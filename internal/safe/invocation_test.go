package safe

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestThreshold(t *testing.T) {
	t.Run("reaches as there are no pauses between the calls", func(t *testing.T) {
		thresholdReached := Threshold(2, time.Hour)

		require.False(t, thresholdReached())
		require.True(t, thresholdReached())
	})

	t.Run("doesn't reach because of pauses between the calls", func(t *testing.T) {
		thresholdReached := Threshold(2, time.Microsecond)

		require.False(t, thresholdReached())
		time.Sleep(time.Millisecond)
		require.False(t, thresholdReached())
	})

	t.Run("always reached for zero values", func(t *testing.T) {
		thresholdReached := Threshold(0, 0)

		require.True(t, thresholdReached())
		time.Sleep(time.Millisecond)
		require.True(t, thresholdReached())
	})
}
