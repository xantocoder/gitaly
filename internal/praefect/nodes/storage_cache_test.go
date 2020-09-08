package nodes

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
)

func TestUpToDateStoragesCache_GetSyncedNodes(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()

		cache := newUpToDateStoragesCache(nil, time.Second, nil)
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

		cacheAccessTotal := mtcs[0]
		assert.Equal(t, "gitaly_praefect_uptodate_storages_cache_access_total", *cacheAccessTotal.Name)
		assert.Len(t, cacheAccessTotal.Metric, 1)
		assert.Len(t, cacheAccessTotal.Metric[0].Label, 2)
		assert.ElementsMatch(t, []string{"vs", "hit"}, []string{*cacheAccessTotal.Metric[0].Label[0].Value, *cacheAccessTotal.Metric[0].Label[1].Value})
		assert.Equal(t, 1.0, *cacheAccessTotal.Metric[0].Counter.Value)

		errorsTotal := mtcs[1]
		assert.Len(t, errorsTotal.Metric, 1)
		assert.Equal(t, 0.0, *errorsTotal.Metric[0].Counter.Value)
	})

	t.Run("cache miss", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()

		rs := datastore.NewMemoryRepositoryStore(map[string][]string{"vs": {"g1", "g2", "g3"}})
		require.NoError(t, rs.SetGeneration(ctx, "vs", "/repo/path", "g1", 1))
		require.NoError(t, rs.SetGeneration(ctx, "vs", "/repo/path", "g2", 1))

		setTriggered := false
		cache := newUpToDateStoragesCache(rs, time.Second, nil)
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
		require.Len(t, mtcs, 2)

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

		errorsTotal := mtcs[1]
		assert.Len(t, errorsTotal.Metric, 1)
		assert.Equal(t, 0.0, *errorsTotal.Metric[0].Counter.Value)
	})

	t.Run("repository store returns error", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()

		logger := testhelper.DiscardTestLogger(t)
		logHook := test.NewLocal(logger)

		rs := datastore.NewMemoryRepositoryStore(map[string][]string{"bad": {"g1", "g2", "g3"}})

		setTriggered := false
		cache := newUpToDateStoragesCache(rs, time.Second, nil)
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
		assert.Equal(t, "gitaly_praefect_uptodate_storages_errors_total", *errorsTotal.Name)
		assert.Len(t, errorsTotal.Metric, 1)
		assert.Equal(t, 1.0, *errorsTotal.Metric[0].Counter.Value)
	})
}

type mockExpirationCache struct {
	SetMethod func(k string, x interface{}, d time.Duration)
	GetMethod func(k string) (interface{}, bool)
}

func (m mockExpirationCache) Set(k string, v interface{}, d time.Duration) {
	m.SetMethod(k, v, d)
}

func (m mockExpirationCache) Get(k string) (interface{}, bool) {
	return m.GetMethod(k)
}
