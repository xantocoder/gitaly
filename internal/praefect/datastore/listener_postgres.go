package datastore

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lib/pq"
	promclient "github.com/prometheus/client_golang/prometheus"
	"gitlab.com/gitlab-org/gitaly/internal/dontpanic"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/datastore/glsql"
	"gitlab.com/gitlab-org/gitaly/internal/safe"
)

// PostgresListenerOpts is a set of configuration options for the PostgreSQL listener.
type PostgresListenerOpts struct {
	Addr                     string
	Channel                  string
	ProcessingPoolSize       int
	PingPeriod               time.Duration
	MinReconnectInterval     time.Duration
	MaxReconnectInterval     time.Duration
	DisconnectThreshold      int
	DisconnectTimeWindow     time.Duration
	ConnectAttemptThreshold  int
	ConnectAttemptTimeWindow time.Duration
}

// DefaultPostgresListenerOpts pre-defined options for PostgreSQL listener.
var DefaultPostgresListenerOpts = PostgresListenerOpts{
	ProcessingPoolSize:       16,
	PingPeriod:               10 * time.Second,
	MinReconnectInterval:     5 * time.Second,
	MaxReconnectInterval:     40 * time.Second,
	DisconnectThreshold:      3,
	DisconnectTimeWindow:     time.Minute,
	ConnectAttemptThreshold:  3,
	ConnectAttemptTimeWindow: time.Minute,
}

// PostgresListener is an implementation based on the PostgreSQL LISTEN/NOTIFY functions.
type PostgresListener struct {
	mtx                            sync.Mutex
	listener                       *pq.Listener
	opts                           PostgresListenerOpts
	wg                             sync.WaitGroup
	closed                         chan struct{}
	err                            error
	reconnectTotal                 *promclient.CounterVec
	extraProcessingGoroutinesTotal promclient.Counter
}

// NewPostgresListener returns a new instance of the listener.
func NewPostgresListener(opts PostgresListenerOpts) (*PostgresListener, error) {
	switch {
	case strings.TrimSpace(opts.Addr) == "":
		return nil, fmt.Errorf("address is invalid: %q", opts.Addr)
	case strings.TrimSpace(opts.Channel) == "":
		return nil, fmt.Errorf("channel is invalid: %q", opts.Channel)
	case opts.ProcessingPoolSize <= 0:
		return nil, fmt.Errorf("invalid processing pool size: %d", opts.ProcessingPoolSize)
	case opts.PingPeriod < 0:
		return nil, fmt.Errorf("invalid ping period: %s", opts.PingPeriod)
	case opts.MinReconnectInterval <= 0:
		return nil, fmt.Errorf("invalid min reconnect period: %s", opts.MinReconnectInterval)
	case opts.MaxReconnectInterval <= 0 || opts.MaxReconnectInterval < opts.MinReconnectInterval:
		return nil, fmt.Errorf("invalid max reconnect period: %s", opts.MaxReconnectInterval)
	}

	return &PostgresListener{
		opts:   opts,
		closed: make(chan struct{}),
		reconnectTotal: promclient.NewCounterVec(
			promclient.CounterOpts{
				Name: "gitaly_praefect_notifications_reconnects_total",
				Help: "Counts amount of reconnects to listen for notification from PostgreSQL",
			},
			[]string{"state"},
		),
		extraProcessingGoroutinesTotal: promclient.NewCounter(
			promclient.CounterOpts{
				Name: "gitaly_praefect_notifications_processing_extra_goroutines_total",
				Help: "Counts amount of extra goroutines spawned to process notification events from PostgreSQL",
			},
		),
	}, nil
}

func (pgl *PostgresListener) Listen(ctx context.Context, handler glsql.ListenHandler) error {
	if err := pgl.initListener(handler); err != nil {
		return err
	}

	notificationsChan := pgl.catchNotifications(ctx, handler)
	pgl.spreadAmongProcessors(notificationsChan, handler)

	pgl.wg.Wait()

	pgl.mtx.Lock()
	defer pgl.mtx.Unlock()

	return pgl.err
}

func listenerEventTypeToString(et pq.ListenerEventType) string {
	switch et {
	case pq.ListenerEventConnected:
		return "connected"
	case pq.ListenerEventDisconnected:
		return "disconnected"
	case pq.ListenerEventReconnected:
		return "reconnected"
	case pq.ListenerEventConnectionAttemptFailed:
		return "connection_attempt_failed"
	}
	return fmt.Sprintf("unknown: %d", et)
}

func (pgl *PostgresListener) initListener(handler glsql.ListenHandler) error {
	pgl.mtx.Lock()
	defer pgl.mtx.Unlock()

	if pgl.listener != nil {
		return fmt.Errorf("already listening channel %q of %q", pgl.opts.Channel, pgl.opts.Addr)
	}
	pgl.err = nil

	disconnectThresholdReached := safe.Threshold(pgl.opts.DisconnectThreshold, pgl.opts.DisconnectTimeWindow)
	connectAttemptThresholdReached := safe.Threshold(pgl.opts.ConnectAttemptThreshold, pgl.opts.ConnectAttemptTimeWindow)

	initialization := int32(1)
	defer atomic.StoreInt32(&initialization, 0)

	connectionLifecycle := func(eventType pq.ListenerEventType, err error) {
		pgl.reconnectTotal.WithLabelValues(listenerEventTypeToString(eventType)).Inc()

		switch eventType {
		case pq.ListenerEventDisconnected:
			dontpanic.Try(handler.Disconnect)
			if disconnectThresholdReached() {
				pgl.close(atomic.LoadInt32(&initialization) == 0, err)
			}
		case pq.ListenerEventConnectionAttemptFailed:
			if connectAttemptThresholdReached() {
				pgl.close(atomic.LoadInt32(&initialization) == 0, err)
			}
		}
	}

	pgl.listener = pq.NewListener(pgl.opts.Addr, pgl.opts.MinReconnectInterval, pgl.opts.MaxReconnectInterval, connectionLifecycle)

	if err := pgl.listener.Listen(pgl.opts.Channel); err != nil {
		return fmt.Errorf("listening for %q: %w", pgl.opts.Channel, err)
	}

	return pgl.err
}

func (pgl *PostgresListener) catchNotifications(ctx context.Context, handler glsql.ListenHandler) <-chan string {
	notificationsChan := make(chan string)

	closed := pgl.closed
	notify := pgl.listener.Notify

	go func() {
		defer close(notificationsChan)

		for {
			select {
			case <-closed:
				return
			case <-ctx.Done():
				pgl.close(true, nil)
				return
			case <-time.After(pgl.opts.PingPeriod):
				if err := pgl.listener.Ping(); err != nil {
					pgl.close(true, err)
				}
			case notification := <-notify:
				if notification == nil {
					continue
				}

				select {
				case notificationsChan <- notification.Extra:
					// handled by the worker from the pool
				default:
					// all workers are busy, so we should start a new goroutine to handle this event
					pgl.wg.Add(1)
					go func() {
						pgl.extraProcessingGoroutinesTotal.Inc()
						defer pgl.wg.Done()
						dontpanic.Try(func() { handler.Notification(notification.Extra) })
					}()
				}
			}
		}
	}()

	return notificationsChan
}

func (pgl *PostgresListener) spreadAmongProcessors(notificationsChan <-chan string, handler glsql.ListenHandler) {
	pgl.wg.Add(pgl.opts.ProcessingPoolSize)
	for i := 0; i < pgl.opts.ProcessingPoolSize; i++ {
		go func() {
			defer pgl.wg.Done()

			for notification := range notificationsChan {
				dontpanic.Try(func() { handler.Notification(notification) })
			}
		}()
	}
}

func (pgl *PostgresListener) close(lock bool, err error) {
	if lock {
		// we should not lock if close was called during initialisation step as lock is acquired by initListener already
		pgl.mtx.Lock()
		defer pgl.mtx.Unlock()
	}

	if pgl.listener == nil {
		return
	}

	pgl.err = err

	uerr := pgl.listener.UnlistenAll()
	cerr := pgl.listener.Close()

	if pgl.err == nil {
		if uerr != nil {
			pgl.err = uerr
		} else {
			pgl.err = cerr
		}
	}

	close(pgl.closed)
	pgl.closed = make(chan struct{})
	pgl.listener = nil
}

func (pgl *PostgresListener) Describe(descs chan<- *promclient.Desc) {
	promclient.DescribeByCollect(pgl, descs)
}

func (pgl *PostgresListener) Collect(metrics chan<- promclient.Metric) {
	pgl.reconnectTotal.Collect(metrics)
	pgl.extraProcessingGoroutinesTotal.Collect(metrics)
}
