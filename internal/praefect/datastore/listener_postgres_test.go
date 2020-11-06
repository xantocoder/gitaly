// +build postgres

package datastore

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/datastore/glsql"
)

func TestNewPostgresListener(t *testing.T) {
	for title, tc := range map[string]struct {
		opts      PostgresListenerOpts
		handler   glsql.ListenHandler
		expErrMsg string
	}{
		"invalid option: address": {
			opts:      PostgresListenerOpts{Addr: ""},
			expErrMsg: "address is invalid",
		},
		"invalid option: channels": {
			opts:      PostgresListenerOpts{Addr: "stub", Channels: nil},
			expErrMsg: "no channels to listen",
		},
		"invalid option: ping period": {
			opts:      PostgresListenerOpts{Addr: "stub", Channels: []string{""}, PingPeriod: -1},
			expErrMsg: "invalid ping period",
		},
		"invalid option: min reconnect period": {
			opts:      PostgresListenerOpts{Addr: "stub", Channels: []string{""}, MinReconnectInterval: 0},
			expErrMsg: "invalid min reconnect period",
		},
		"invalid option: max reconnect period": {
			opts:      PostgresListenerOpts{Addr: "stub", Channels: []string{""}, MinReconnectInterval: time.Second, MaxReconnectInterval: time.Millisecond},
			expErrMsg: "invalid max reconnect period",
		},
	} {
		t.Run(title, func(t *testing.T) {
			pgl, err := NewPostgresListener(tc.opts, nil)
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

type mockListenHandler struct {
	OnNotification func(glsql.Notification)
	OnDisconnect   func(error)
	OnConnected    func()
}

func (mlh mockListenHandler) Notification(n glsql.Notification) {
	if mlh.OnNotification != nil {
		mlh.OnNotification(n)
	}
}

func (mlh mockListenHandler) Disconnect(err error) {
	if mlh.OnDisconnect != nil {
		mlh.OnDisconnect(err)
	}
}

func (mlh mockListenHandler) Connected() {
	if mlh.OnConnected != nil {
		mlh.OnConnected()
	}
}

func TestPostgresListener_Listen(t *testing.T) {
	db := getDB(t)

	newOpts := func() PostgresListenerOpts {
		opts := DefaultPostgresListenerOpts
		opts.Addr = getDBConfig(t).ToPQString(true)
		opts.MinReconnectInterval = time.Nanosecond
		opts.MaxReconnectInterval = time.Minute
		return opts
	}

	newChannel := func(i int) func() string {
		return func() string {
			i++
			return fmt.Sprintf("channel_%d", i)
		}
	}(0)

	notifyListener := func(t *testing.T, channelName, payload string) {
		t.Helper()

		_, err := db.Exec(fmt.Sprintf(`NOTIFY %s, '%s'`, channelName, payload))
		assert.NoError(t, err)
	}

	listenNotify := func(t *testing.T, opts PostgresListenerOpts, numNotifiers int, payloads []string) (*PostgresListener, []string) {
		t.Helper()

		start := make(chan struct{})
		done := make(chan struct{})
		defer func() { <-done }()

		numResults := len(payloads) * numNotifiers
		allReceivedChan := make(chan struct{})

		go func() {
			defer close(done)

			<-start

			var wg sync.WaitGroup
			for i := 0; i < numNotifiers; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for _, payload := range payloads {
						for _, channel := range opts.Channels {
							notifyListener(t, channel, payload)
						}
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
		callback := func(idx int) func(n glsql.Notification) {
			return func(n glsql.Notification) {
				idx++
				result[idx] = n.Payload
				if idx+1 == numResults {
					close(allReceivedChan)
				}
			}
		}(-1)

		handler := mockListenHandler{OnNotification: callback, OnConnected: func() { close(start) }}
		pgl, err := NewPostgresListener(opts, handler)
		require.NoError(t, err)

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

	waitFor := func(t *testing.T, c <-chan struct{}, d time.Duration) {
		t.Helper()

		select {
		case <-time.After(d):
			require.FailNow(t, "it takes too long")
		case <-c:
			// proceed
		}
	}

	t.Run("single handler and single notifier", func(t *testing.T) {
		opts := newOpts()
		opts.Channels = []string{newChannel()}

		payloads := []string{"this", "is", "a", "payload"}

		listener, result := listenNotify(t, opts, 1, payloads)
		defer func() { require.NoError(t, listener.Close()) }()
		require.Equal(t, payloads, result)
	})

	t.Run("single handler and multiple notifiers", func(t *testing.T) {
		opts := newOpts()
		opts.Channels = []string{newChannel()}

		numNotifiers := 10

		payloads := []string{"this", "is", "a", "payload"}
		var expResult []string
		for i := 0; i < numNotifiers; i++ {
			expResult = append(expResult, payloads...)
		}

		listener, result := listenNotify(t, opts, numNotifiers, payloads)
		defer func() { require.NoError(t, listener.Close()) }()
		require.ElementsMatch(t, expResult, result, "there must be no additional data, only expected")
	})

	t.Run("multiple channels", func(t *testing.T) {
		opts := newOpts()
		channel1 := newChannel()
		channel2 := newChannel()
		opts.Channels = []string{channel1, channel2}

		start := make(chan struct{})
		resultChans := map[string]chan glsql.Notification{
			channel1: make(chan glsql.Notification, 1),
			channel2: make(chan glsql.Notification, 1),
		}
		handler := mockListenHandler{
			OnNotification: func(n glsql.Notification) { resultChans[n.Channel] <- n },
			OnConnected:    func() { close(start) },
		}

		listener, err := NewPostgresListener(opts, handler)
		require.NoError(t, err)
		defer func() { require.NoError(t, listener.Close()) }()

		waitFor(t, start, time.Minute)

		payloads := []string{"1", "2", "3"}
		for _, payload := range payloads {
			for _, channel := range opts.Channels {
				notifyListener(t, channel, channel+":"+payload)
			}
		}

		tooLong := time.After(time.Minute)
		var notifications []glsql.Notification
		for i := 0; i < len(opts.Channels)*len(payloads); i++ {
			select {
			case <-tooLong:
				require.FailNow(t, "no notifications for too long")
			case p := <-resultChans[channel1]:
				notifications = append(notifications, p)
			case p := <-resultChans[channel2]:
				notifications = append(notifications, p)
			}
		}

		require.Equal(t, []glsql.Notification{
			{Channel: channel1, Payload: channel1 + ":1"}, {Channel: channel2, Payload: channel2 + ":1"},
			{Channel: channel1, Payload: channel1 + ":2"}, {Channel: channel2, Payload: channel2 + ":2"},
			{Channel: channel1, Payload: channel1 + ":3"}, {Channel: channel2, Payload: channel2 + ":3"},
		}, notifications)
	})

	t.Run("invalid connection", func(t *testing.T) {
		opts := newOpts()
		opts.Addr = "invalid-address"
		opts.Channels = []string{"stub"}

		_, err := NewPostgresListener(opts, mockListenHandler{})
		require.Error(t, err)
		require.Regexp(t, "^connect: .*invalid-address.*", err.Error())
	})

	t.Run("channel used more then once", func(t *testing.T) {
		opts := newOpts()
		opts.Channels = []string{"stub1", "stub2", "stub1"}

		_, err := NewPostgresListener(opts, mockListenHandler{})
		require.True(t, errors.Is(err, pq.ErrChannelAlreadyOpen), err)
	})

	t.Run("connection interruption", func(t *testing.T) {
		opts := newOpts()
		opts.Channels = []string{newChannel()}

		connected := make(chan struct{}, 1)
		handler := mockListenHandler{OnConnected: func() { connected <- struct{}{} }}

		listener, err := NewPostgresListener(opts, handler)
		require.NoError(t, err)

		waitFor(t, connected, time.Minute)
		disconnectListener(t, opts.Channels[0])
		waitFor(t, connected, time.Minute)
		require.NoError(t, listener.Close())

		err = testutil.CollectAndCompare(listener, strings.NewReader(`
			# HELP gitaly_praefect_notifications_reconnects_total Counts amount of reconnects to listen for notification from PostgreSQL
			# TYPE gitaly_praefect_notifications_reconnects_total counter
			gitaly_praefect_notifications_reconnects_total{state="connected"} 1
			gitaly_praefect_notifications_reconnects_total{state="disconnected"} 1
			gitaly_praefect_notifications_reconnects_total{state="reconnected"} 1
		`))
		require.NoError(t, err)
	})

	t.Run("persisted connection interruption", func(t *testing.T) {
		opts := newOpts()
		opts.Channels = []string{newChannel()}

		connected := make(chan struct{}, 1)
		disconnected := make(chan struct{}, 1)
		handler := mockListenHandler{
			OnConnected: func() { connected <- struct{}{} },
			OnDisconnect: func(err error) {
				assert.Error(t, err, "disconnect event should always receive non-nil error")
				disconnected <- struct{}{}
			},
		}

		listener, err := NewPostgresListener(opts, handler)
		require.NoError(t, err)

		for i := 0; i < 3; i++ {
			waitFor(t, connected, time.Minute)
			disconnectListener(t, opts.Channels[0])
			waitFor(t, disconnected, time.Minute)
		}

		// this additional step is required to have exactly 3 "reconnected" metric value, otherwise it could
		// be 2 or 3 - it depends if it was quick enough to re-establish a new connection or not.
		waitFor(t, connected, time.Minute)

		require.NoError(t, listener.Close())

		err = testutil.CollectAndCompare(listener, strings.NewReader(`
			# HELP gitaly_praefect_notifications_reconnects_total Counts amount of reconnects to listen for notification from PostgreSQL
			# TYPE gitaly_praefect_notifications_reconnects_total counter
			gitaly_praefect_notifications_reconnects_total{state="connected"} 1
			gitaly_praefect_notifications_reconnects_total{state="disconnected"} 3
			gitaly_praefect_notifications_reconnects_total{state="reconnected"} 3
		`))
		require.NoError(t, err)
	})
}

func TestPostgresListener_Listen_repositories_delete(t *testing.T) {
	db := getDB(t)

	const channel = "repositories_updates"

	testListener(
		t,
		"repositories_updates",
		func(t *testing.T) {
			_, err := db.DB.Exec(`
				INSERT INTO repositories
				VALUES ('praefect-1', '/path/to/repo/1', 1),
					('praefect-1', '/path/to/repo/2', 1),
					('praefect-1', '/path/to/repo/3', 0)`)
			require.NoError(t, err)
		},
		func(t *testing.T) {
			_, err := db.DB.Exec(`DELETE FROM repositories WHERE generation > 0`)
			require.NoError(t, err)
		},
		func(t *testing.T, n glsql.Notification) {
			require.Equal(t, channel, n.Channel)
			require.JSONEq(t, `
				{
					"old": [
						{"virtual_storage":"praefect-1","relative_path":"/path/to/repo/1","generation":1,"primary":null},
						{"virtual_storage":"praefect-1","relative_path":"/path/to/repo/2","generation":1,"primary":null}
					],
					"new" : null
				}`,
				n.Payload,
			)
		},
	)
}

func TestPostgresListener_Listen_storage_repositories_insert(t *testing.T) {
	db := getDB(t)

	const channel = "storage_repositories_updates"

	testListener(
		t,
		channel,
		func(t *testing.T) {},
		func(t *testing.T) {
			_, err := db.DB.Exec(`
				INSERT INTO storage_repositories
				VALUES ('praefect-1', '/path/to/repo', 'gitaly-1', 0),
					('praefect-1', '/path/to/repo', 'gitaly-2', 0)`,
			)
			require.NoError(t, err)
		},
		func(t *testing.T, n glsql.Notification) {
			require.Equal(t, channel, n.Channel)
			require.JSONEq(t, `
				{
					"old":null,
					"new":[
						{"virtual_storage":"praefect-1","relative_path":"/path/to/repo","storage":"gitaly-1","generation":0},
						{"virtual_storage":"praefect-1","relative_path":"/path/to/repo","storage":"gitaly-2","generation":0}
					]
				}`,
				n.Payload,
			)
		},
	)
}

func TestPostgresListener_Listen_storage_repositories_update(t *testing.T) {
	db := getDB(t)

	const channel = "storage_repositories_updates"

	testListener(
		t,
		channel,
		func(t *testing.T) {
			_, err := db.DB.Exec(`INSERT INTO storage_repositories VALUES ('praefect-1', '/path/to/repo', 'gitaly-1', 0)`)
			require.NoError(t, err)
		},
		func(t *testing.T) {
			_, err := db.DB.Exec(`UPDATE storage_repositories SET generation = generation + 1`)
			require.NoError(t, err)
		},
		func(t *testing.T, n glsql.Notification) {
			require.Equal(t, channel, n.Channel)
			require.JSONEq(t, `
				{
					"old" : [{"virtual_storage":"praefect-1","relative_path":"/path/to/repo","storage":"gitaly-1","generation":0}],
					"new" : [{"virtual_storage":"praefect-1","relative_path":"/path/to/repo","storage":"gitaly-1","generation":1}]
				}`,
				n.Payload,
			)
		},
	)
}

func TestPostgresListener_Listen_storage_repositories_delete(t *testing.T) {
	db := getDB(t)

	const channel = "storage_repositories_updates"

	testListener(
		t,
		channel,
		func(t *testing.T) {
			_, err := db.DB.Exec(`
				INSERT INTO storage_repositories (virtual_storage, relative_path, storage, generation)
				VALUES ('praefect-1', '/path/to/repo', 'gitaly-1', 0)`,
			)
			require.NoError(t, err)
		},
		func(t *testing.T) {
			_, err := db.DB.Exec(`DELETE FROM storage_repositories`)
			require.NoError(t, err)
		},
		func(t *testing.T, n glsql.Notification) {
			require.Equal(t, channel, n.Channel)
			require.JSONEq(t, `
				{
					"old" : [{"virtual_storage":"praefect-1","relative_path":"/path/to/repo","storage":"gitaly-1","generation":0}],
					"new" : null
				}`,
				n.Payload,
			)
		},
	)
}

func testListener(t *testing.T, channel string, setup func(t *testing.T), trigger func(t *testing.T), verifier func(t *testing.T, notification glsql.Notification)) {
	setup(t)

	readyChan := make(chan struct{})
	receivedChan := make(chan struct{})
	var notification glsql.Notification

	callback := func(n glsql.Notification) {
		select {
		case <-receivedChan:
			return
		default:
			notification = n
			close(receivedChan)
		}
	}

	opts := DefaultPostgresListenerOpts
	opts.Addr = getDBConfig(t).ToPQString(true)
	opts.Channels = []string{channel}

	handler := mockListenHandler{OnNotification: callback, OnConnected: func() { close(readyChan) }}

	pgl, err := NewPostgresListener(opts, handler)
	require.NoError(t, err)
	defer pgl.Close()

	select {
	case <-time.After(time.Second):
		require.FailNow(t, "no connection for too long period")
	case <-readyChan:
	}

	trigger(t)

	select {
	case <-time.After(time.Second):
		require.FailNow(t, "no notifications for too long period")
	case <-receivedChan:
	}

	verifier(t, notification)
}
