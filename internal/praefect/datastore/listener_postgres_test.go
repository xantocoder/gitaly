// +build postgres

package datastore

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	promclient "github.com/prometheus/client_golang/prometheus"
	promclientgo "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
)

func TestNewPostgresListener(t *testing.T) {
	for title, tc := range map[string]struct {
		opts      PostgresListenerOpts
		expErrMsg string
	}{
		"all set": {
			opts: PostgresListenerOpts{
				Addr:                 "stub",
				Channel:              "sting",
				ProcessingPoolSize:   1,
				MinReconnectInterval: time.Second,
				MaxReconnectInterval: time.Minute,
			},
		},
		"invalid option: address": {
			opts:      PostgresListenerOpts{Addr: ""},
			expErrMsg: "address is invalid",
		},
		"invalid option: channel": {
			opts:      PostgresListenerOpts{Addr: "stub", Channel: "  "},
			expErrMsg: "channel is invalid",
		},
		"invalid option: processing pool size": {
			opts:      PostgresListenerOpts{Addr: "stub", Channel: "stub", ProcessingPoolSize: 0},
			expErrMsg: "invalid processing pool size",
		},
		"invalid option: ping period": {
			opts:      PostgresListenerOpts{Addr: "stub", Channel: "stub", ProcessingPoolSize: 1, PingPeriod: -1},
			expErrMsg: "invalid ping period",
		},
		"invalid option: min reconnect period": {
			opts:      PostgresListenerOpts{Addr: "stub", Channel: "stub", ProcessingPoolSize: 1, MinReconnectInterval: 0},
			expErrMsg: "invalid min reconnect period",
		},
		"invalid option: max reconnect period": {
			opts:      PostgresListenerOpts{Addr: "stub", Channel: "stub", ProcessingPoolSize: 1, MinReconnectInterval: time.Second, MaxReconnectInterval: time.Millisecond},
			expErrMsg: "invalid max reconnect period",
		},
	} {
		t.Run(title, func(t *testing.T) {
			pgl, err := NewPostgresListener(tc.opts)
			if tc.expErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, pgl)
		})
	}
}

func TestPostgresListener_Listen(t *testing.T) {
	db := getDB(t)

	newOpts := func() PostgresListenerOpts {
		opts := DefaultPostgresListenerOpts
		opts.Addr = getDBConfig(t).ToPQString(false)
		opts.Channel = fmt.Sprintf("channel_%d", time.Now().UnixNano())
		opts.MinReconnectInterval = time.Nanosecond
		opts.MaxReconnectInterval = time.Second
		return opts
	}

	notifyListener := func(t *testing.T, channelName, payload string) {
		t.Helper()

		_, err := db.Exec(fmt.Sprintf(`NOTIFY %s, '%s'`, channelName, payload))
		assert.NoError(t, err)
	}

	listenNotify := func(t *testing.T, opts PostgresListenerOpts, numNotifiers int, payloads []string) (*PostgresListener, []string) {
		t.Helper()

		pgl, err := NewPostgresListener(opts)

		require.NoError(t, err)

		ctx, cancel := testhelper.Context()
		defer cancel()

		numResults := len(payloads) * numNotifiers
		allReceivedChan := make(chan struct{})

		go func() {
			defer cancel()

			time.Sleep(100 * time.Millisecond)

			var wg sync.WaitGroup
			wg.Add(numNotifiers)
			for i := 0; i < numNotifiers; i++ {
				go func() {
					defer wg.Done()

					for _, payload := range payloads {
						notifyListener(t, opts.Channel, payload)
					}
				}()
			}
			wg.Wait()

			select {
			case <-time.After(time.Second):
				assert.FailNow(t, "notification propagation takes too long")
			case <-allReceivedChan:
			}
		}()

		result := make([]string, numResults)
		idx := int32(-1)
		callback := func(payload string) {
			i := int(atomic.AddInt32(&idx, 1))
			result[i] = payload
			if i+1 == numResults {
				close(allReceivedChan)
			}
		}

		require.NoError(t, pgl.Listen(ctx, callback))

		return pgl, result
	}

	disconnectListener := func(t *testing.T, channelName string) {
		t.Helper()

		q := `SELECT PG_TERMINATE_BACKEND(pid) FROM PG_STAT_ACTIVITY WHERE datname = $1 AND query = $2`
		res, err := db.Exec(q, databaseName, fmt.Sprintf("LISTEN %q", channelName))
		if assert.NoError(t, err) {
			affected, err := res.RowsAffected()
			assert.NoError(t, err)
			assert.EqualValues(t, 1, affected)
		}
	}

	readMetrics := func(t *testing.T, col promclient.Collector) []promclientgo.Metric {
		t.Helper()

		metricsChan := make(chan promclient.Metric, 16)
		col.Collect(metricsChan)
		close(metricsChan)
		var metric []promclientgo.Metric
		for m := range metricsChan {
			var mtc promclientgo.Metric
			assert.NoError(t, m.Write(&mtc))
			metric = append(metric, mtc)
		}
		return metric
	}

	t.Run("single processor and single notifier", func(t *testing.T) {
		opts := newOpts()
		opts.ProcessingPoolSize = 1

		payloads := []string{"this", "is", "a", "payload"}

		listener, result := listenNotify(t, opts, 1, payloads)
		require.Equal(t, payloads, result)

		metrics := readMetrics(t, listener.reconnectTotal)
		require.Len(t, metrics, 1)
		require.Len(t, metrics[0].Label, 1)
		require.Equal(t, "state", *metrics[0].Label[0].Name)
		require.Equal(t, "connected", *metrics[0].Label[0].Value)
		require.GreaterOrEqual(t, *metrics[0].Counter.Value, 1.0)
	})

	t.Run("single processor and multiple notifiers", func(t *testing.T) {
		opts := newOpts()
		opts.ProcessingPoolSize = 1

		numNotifiers := 10

		payloads := []string{"this", "is", "a", "payload"}
		var expResult []string
		for i := 0; i < numNotifiers; i++ {
			expResult = append(expResult, payloads...)
		}

		listener, result := listenNotify(t, opts, numNotifiers, payloads)
		assert.ElementsMatch(t, expResult, result, "there must be no additional data, only expected")

		metrics := readMetrics(t, listener.extraProcessingGoroutinesTotal)
		require.Len(t, metrics, 1)
		require.Greater(t, *metrics[0].Counter.Value, 1.0, "single goroutine can't process all notifications, so it is expected additional goroutines to be created")
	})

	t.Run("multiple processors and multiple notifiers", func(t *testing.T) {
		opts := newOpts()
		opts.ProcessingPoolSize = 10

		numNotifiers := 10

		payloads := []string{"this", "is", "a", "payload"}
		var expResult []string
		for i := 0; i < numNotifiers; i++ {
			expResult = append(expResult, payloads...)
		}

		_, result := listenNotify(t, opts, numNotifiers, payloads)
		assert.ElementsMatch(t, expResult, result, "there must be no additional data, only expected")
	})

	t.Run("re-listen", func(t *testing.T) {
		opts := newOpts()
		listener, result := listenNotify(t, opts, 1, []string{"1"})
		require.Equal(t, []string{"1"}, result)

		ctx, cancel := testhelper.Context()

		var connected int32

		errCh := make(chan error, 1)
		go func() {
			errCh <- listener.Listen(ctx, func(payload string) {
				atomic.StoreInt32(&connected, 1)
				assert.Equal(t, "2", payload)
			})
		}()

		for atomic.LoadInt32(&connected) == 0 {
			notifyListener(t, opts.Channel, "2")
		}

		cancel()
		err := <-errCh
		require.NoError(t, err)
	})

	t.Run("already listening", func(t *testing.T) {
		opts := newOpts()

		listener, err := NewPostgresListener(opts)
		require.NoError(t, err)

		ctx, cancel := testhelper.Context()
		defer cancel()

		var connected int32

		errCh := make(chan error, 1)
		go func() {
			errCh <- listener.Listen(ctx, func(payload string) {
				atomic.StoreInt32(&connected, 1)
				assert.Equal(t, "2", payload)
			})
		}()

		for atomic.LoadInt32(&connected) == 0 {
			notifyListener(t, opts.Channel, "2")
		}

		err = listener.Listen(ctx, func(payload string) {})
		require.Error(t, err)
		require.Equal(t, fmt.Sprintf(`already listening channel %q of %q`, opts.Channel, opts.Addr), err.Error())
	})

	t.Run("invalid connection", func(t *testing.T) {
		opts := newOpts()
		opts.Addr = "invalid-address"

		listener, err := NewPostgresListener(opts)
		require.NoError(t, err)

		ctx, cancel := testhelper.Context()
		defer cancel()

		err = listener.Listen(ctx, func(string) {
			assert.FailNow(t, "no notifications expected to be received")
		})
		require.Error(t, err, "it should not be possible to start listening on invalid connection")
	})

	t.Run("connection interruption", func(t *testing.T) {
		opts := newOpts()
		listener, err := NewPostgresListener(opts)
		require.NoError(t, err)

		ctx, cancel := testhelper.Context()
		defer cancel()

		var connected int32

		errChan := make(chan error, 1)
		go func() {
			errChan <- listener.Listen(ctx, func(string) {
				atomic.StoreInt32(&connected, 1)
			})
		}()

		for atomic.LoadInt32(&connected) == 0 {
			notifyListener(t, opts.Channel, "")
		}

		disconnectListener(t, opts.Channel)
		atomic.StoreInt32(&connected, 0)

		for atomic.LoadInt32(&connected) == 0 {
			notifyListener(t, opts.Channel, "")
		}

		cancel()
		require.NoError(t, <-errChan)
	})

	t.Run("persisted connection interruption", func(t *testing.T) {
		opts := newOpts()
		listener, err := NewPostgresListener(opts)
		require.NoError(t, err)

		ctx, cancel := testhelper.Context()
		defer cancel()

		var connected int32

		errChan := make(chan error, 1)
		go func() {
			errChan <- listener.Listen(ctx, func(string) {
				atomic.StoreInt32(&connected, 1)
			})
		}()

		for i := 0; i < 3; i++ {
			for atomic.LoadInt32(&connected) == 0 {
				notifyListener(t, opts.Channel, "")
			}

			disconnectListener(t, opts.Channel)
			atomic.StoreInt32(&connected, 0)
		}

		require.Error(t, <-errChan, "it should not be possible to start listening on invalid connection")

		metrics := readMetrics(t, listener.reconnectTotal)
		for _, metric := range metrics {
			switch *metric.Label[0].Value {
			case "connected":
				require.GreaterOrEqual(t, *metric.Counter.Value, 1.0)
			case "disconnected":
				require.GreaterOrEqual(t, *metric.Counter.Value, 3.0)
			case "reconnected":
				require.GreaterOrEqual(t, *metric.Counter.Value, 2.0)
			}
		}
	})
}
