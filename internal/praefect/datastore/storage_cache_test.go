package datastore

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	promclient "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
)

func TestUpToDateStoragesCache_GetSyncedNodes(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()

		cache := NewUpToDateStoragesCache(testhelper.NewTestLogger(t), nil, time.Second, nil)
		cache.cacheByVirtualStorage = map[string]expirationCache{
			"vs": mockExpirationCache{
				SetMethod: func(k string, v interface{}, d time.Duration) {
					require.FailNow(t, "should not set any value: %v", v)
				},
				GetMethod: func(k string) (interface{}, bool) {
					require.Equal(t, "/repo/path", k)
					return []string{"g2", "g3"}, true
				},
			},
		}

		promRegistry := prometheus.NewRegistry()
		require.NoError(t, promRegistry.Register(cache))

		storages := cache.GetSyncedNodes(ctx, nil, "vs", "/repo/path", "g1")
		require.Equal(t, []string{"g2", "g3"}, storages)

		mtcs, err := promRegistry.Gather()
		require.NoError(t, err)
		require.Len(t, mtcs, 1)

		cacheAccessTotal := mtcs[0]
		assert.Equal(t, "gitaly_praefect_uptodate_storages_cache_access_total", *cacheAccessTotal.Name)
		assert.Len(t, cacheAccessTotal.Metric, 1)
		assert.Len(t, cacheAccessTotal.Metric[0].Label, 2)
		assert.ElementsMatch(t, []string{"vs", "hit"}, []string{*cacheAccessTotal.Metric[0].Label[0].Value, *cacheAccessTotal.Metric[0].Label[1].Value})
		assert.Equal(t, 1.0, *cacheAccessTotal.Metric[0].Counter.Value)
	})

	t.Run("cache miss", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()

		rs := NewMemoryRepositoryStore(map[string][]string{"vs": {"g1", "g2", "g3"}})
		require.NoError(t, rs.SetGeneration(ctx, "vs", "/repo/path", "g1", 1))
		require.NoError(t, rs.SetGeneration(ctx, "vs", "/repo/path", "g2", 1))

		setTriggered := false
		cache := NewUpToDateStoragesCache(testhelper.NewTestLogger(t), rs, time.Second, nil)
		cache.cacheByVirtualStorage = map[string]expirationCache{
			"vs": mockExpirationCache{
				SetMethod: func(k string, v interface{}, d time.Duration) {
					require.Equal(t, "/repo/path", k)
					require.ElementsMatch(t, []string{"g1", "g2"}, v)
					require.Equal(t, time.Second, d)
					setTriggered = true
				},
				GetMethod: func(k string) (interface{}, bool) {
					require.Equal(t, "/repo/path", k)
					return nil, false
				},
			},
		}

		promRegistry := prometheus.NewRegistry()
		require.NoError(t, promRegistry.Register(cache))

		storages := cache.GetSyncedNodes(ctx, nil, "vs", "/repo/path", "g1")
		require.ElementsMatch(t, []string{"g1", "g2"}, storages)
		require.True(t, setTriggered)

		mtcs, err := promRegistry.Gather()
		require.NoError(t, err)
		require.Len(t, mtcs, 1)

		cacheAccessTotal := mtcs[0]
		assert.Len(t, cacheAccessTotal.Metric, 2)

		cacheAccessTotalMetricsMiss := cacheAccessTotal.Metric[0]
		assert.Len(t, cacheAccessTotalMetricsMiss.Label, 2)

		assert.ElementsMatch(t, []string{"vs", "miss"}, []string{*cacheAccessTotalMetricsMiss.Label[0].Value, *cacheAccessTotalMetricsMiss.Label[1].Value})
		assert.Equal(t, 1.0, *cacheAccessTotalMetricsMiss.Counter.Value)

		cacheAccessTotalMetricsPopulate := cacheAccessTotal.Metric[1]
		assert.Len(t, cacheAccessTotalMetricsPopulate.Label, 2)

		assert.ElementsMatch(t, []string{"vs", "populate"}, []string{*cacheAccessTotalMetricsPopulate.Label[0].Value, *cacheAccessTotalMetricsPopulate.Label[1].Value})
		assert.Equal(t, 1.0, *cacheAccessTotalMetricsPopulate.Counter.Value)
	})

	t.Run("repository store returns an error", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()

		logger := testhelper.DiscardTestLogger(t)
		logHook := test.NewLocal(logger)

		rs := NewMemoryRepositoryStore(map[string][]string{"bad": {"g1", "g2", "g3"}})

		setTriggered := false
		cache := NewUpToDateStoragesCache(testhelper.NewTestLogger(t), rs, time.Second, nil)
		cache.cacheByVirtualStorage = map[string]expirationCache{
			"vs": mockExpirationCache{
				SetMethod: func(k string, v interface{}, d time.Duration) {
					setTriggered = true
				},
				GetMethod: func(k string) (interface{}, bool) {
					return nil, false
				},
			},
		}

		promRegistry := prometheus.NewRegistry()
		require.NoError(t, promRegistry.Register(cache))

		storages := cache.GetSyncedNodes(ctx, logger, "vs", "/repo/path", "g1")
		require.ElementsMatch(t, []string{"g1"}, storages)
		require.True(t, setTriggered)
		require.Len(t, logHook.AllEntries(), 1)
		require.Equal(t, "get up to date secondaries", logHook.LastEntry().Message)
		require.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)

		mtcs, err := promRegistry.Gather()
		require.NoError(t, err)
		require.Len(t, mtcs, 2)

		errorsTotal := mtcs[1]
		require.Len(t, errorsTotal.Metric, 1)
		verifyMetric(t, errorsTotal.Metric[0], 1.0, map[string]string{"type": "sql_retrieve"})
	})

	t.Run("occurrence of cache invalidation", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()

		rs := NewMemoryRepositoryStore(map[string][]string{"vs": {"g1", "g2"}})
		require.NoError(t, rs.SetGeneration(ctx, "vs", "/repo/path/1", "g1", 1))
		require.NoError(t, rs.SetGeneration(ctx, "vs", "/repo/path/1", "g2", 1))

		setTriggered := map[string]int{}
		deleteTriggered := map[string]int{}

		logger := testhelper.NewTestLogger(t)
		hook := test.NewLocal(logger)
		cache := NewUpToDateStoragesCache(logger, rs, time.Second, nil)
		cache.cacheByVirtualStorage = map[string]expirationCache{
			"vs": mockExpirationCache{
				SetMethod: func(k string, v interface{}, d time.Duration) {
					setTriggered[k]++
				},
				GetMethod: func(k string) (interface{}, bool) {
					return nil, false
				},
				DeleteMethod: func(k string) {
					deleteTriggered[k]++
				},
			},
		}

		promRegistry := prometheus.NewRegistry()
		require.NoError(t, promRegistry.Register(cache))

		cache.Notification(``) // invalid notification payload to verify it would be properly handled
		cache.Notification(`
			{
				"old":[
					{"virtual_storage": "vs", "relative_path": "/repo/path/1"},
					{"virtual_storage": "bad", "relative_path": "/repo/path/1"}
				],
				"new":[{"virtual_storage": "vs", "relative_path": "/repo/path/2"}]
			}`,
		)

		require.Equal(t, map[string]int{"/repo/path/1": 1, "/repo/path/2": 1}, deleteTriggered)
		logEntries := hook.AllEntries()
		require.Len(t, logEntries, 2)
		assert.Equal(t, logrus.Fields{
			"component": "up_to_date_storages_cache",
			"error":     io.EOF,
		}, logEntries[0].Data)
		assert.Equal(t, "received payload can't be processed", logEntries[0].Message)

		assert.Equal(t, logrus.Fields{
			"component": "up_to_date_storages_cache",
			"entry":     notificationEntry{VirtualStorage: "bad", RelativePath: "/repo/path/1"},
			"error":     ErrVirtualStorageNotExist,
		}, logEntries[1].Data)
		assert.Equal(t, "can't be removed from the cache", logEntries[1].Message)

		storages := cache.GetSyncedNodes(ctx, nil, "vs", "/repo/path/1", "g1")
		require.ElementsMatch(t, []string{"g1", "g2"}, storages)
		require.Equal(t, map[string]int{"/repo/path/1": 1}, setTriggered)

		mtcs, err := promRegistry.Gather()
		require.NoError(t, err)
		require.Len(t, mtcs, 2)

		cacheAccessTotal := mtcs[0]
		assert.Len(t, cacheAccessTotal.Metric, 3)
		verifyMetric(t, cacheAccessTotal.Metric[0], 2.0, map[string]string{"virtual_storage": "vs", "type": "evict"})
		verifyMetric(t, cacheAccessTotal.Metric[1], 1.0, map[string]string{"virtual_storage": "vs", "type": "miss"})
		verifyMetric(t, cacheAccessTotal.Metric[2], 1.0, map[string]string{"virtual_storage": "vs", "type": "populate"})

		errorsTotal := mtcs[1]
		require.Len(t, errorsTotal.Metric, 2)
		verifyMetric(t, errorsTotal.Metric[0], 1.0, map[string]string{"type": "cache_entry_removal"})
		verifyMetric(t, errorsTotal.Metric[1], 1.0, map[string]string{"type": "notification_decode"})
	})
}

func TestDumbStorageProvider_GetSyncedNodes(t *testing.T) {
	ctx, cancel := testhelper.Context()
	defer cancel()

	t.Run("found", func(t *testing.T) {
		rs := NewMemoryRepositoryStore(map[string][]string{"vs": {"g1", "g2"}})
		require.NoError(t, rs.SetGeneration(ctx, "vs", "/repo/path/1", "g1", 1))
		require.NoError(t, rs.SetGeneration(ctx, "vs", "/repo/path/1", "g2", 1))

		provider := DumbStorageProvider{RepositoryStore: rs}
		nodes := provider.GetSyncedNodes(ctx, nil, "vs", "/repo/path/1", "g1")
		require.ElementsMatch(t, []string{"g1", "g2"}, nodes)
	})

	t.Run("none found", func(t *testing.T) {
		logger := testhelper.NewTestLogger(t)
		provider := DumbStorageProvider{RepositoryStore: NewMemoryRepositoryStore(nil)}
		nodes := provider.GetSyncedNodes(ctx, logger, "vs", "/repo/path/1", "g1")
		require.Equal(t, []string{"g1"}, nodes)
	})

	t.Run("error", func(t *testing.T) {
		logger := testhelper.NewTestLogger(t)
		hook := test.NewLocal(logger)

		provider := DumbStorageProvider{RepositoryStore: mockRepositoryStore{}}
		nodes := provider.GetSyncedNodes(ctx, logger, "vs", "/repo/path/1", "g1")
		require.Equal(t, []string{"g1"}, nodes)

		logEntries := hook.AllEntries()
		require.Len(t, logEntries, 1)
		assert.Equal(t, logrus.Fields{"error": assert.AnError}, logEntries[0].Data)
		assert.Equal(t, "dumb impl: get up to date secondaries", logEntries[0].Message)
	})
}

type mockRepositoryStore struct{ RepositoryStore }

func (mockRepositoryStore) GetConsistentSecondaries(_ context.Context, _, _, _ string) (map[string]struct{}, error) {
	return nil, assert.AnError
}

func verifyMetric(t *testing.T, metric *promclient.Metric, val float64, labels map[string]string) {
	t.Helper()

	assert.Len(t, metric.Label, len(labels))
	actual := make(map[string]string, len(labels))
	for _, lbl := range metric.Label {
		actual[*lbl.Name] = *lbl.Value
	}
	assert.Equal(t, actual, labels)
	assert.Equal(t, val, *metric.Counter.Value)
}

type mockExpirationCache struct {
	SetMethod    func(k string, x interface{}, d time.Duration)
	GetMethod    func(k string) (interface{}, bool)
	DeleteMethod func(k string)
}

func (m mockExpirationCache) Set(k string, v interface{}, d time.Duration) {
	m.SetMethod(k, v, d)
}

func (m mockExpirationCache) Get(k string) (interface{}, bool) {
	return m.GetMethod(k)
}

func (m mockExpirationCache) Delete(k string) {
	m.DeleteMethod(k)
}
