package nodes

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/datastore"
	"gitlab.com/gitlab-org/gitaly/internal/testhelper"
)

func TestUpToDateStoragesCache_GetSyncedNodes(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()

		cache := upToDateStoragesCache{
			exp: time.Second,
			cacheByVirtualStorage: map[string]expirationCache{
				"vs": mockExpirationCache{
					SetMethod: func(k string, v interface{}, d time.Duration) {
						require.FailNow(t, "should not set any value: %v", v)
					},
					GetMethod: func(k string) (interface{}, bool) {
						require.Equal(t, "/repo/path", k)
						return []string{"g2", "g3"}, true
					},
				},
			},
		}

		storages := cache.GetSyncedNodes(ctx, nil, "vs", "/repo/path", "g1")
		require.Equal(t, []string{"g2", "g3"}, storages)
	})

	t.Run("cache miss", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()

		rs := datastore.NewMemoryRepositoryStore(map[string][]string{"vs": {"g1", "g2", "g3"}})
		require.NoError(t, rs.SetGeneration(ctx, "vs", "/repo/path", "g1", 1))
		require.NoError(t, rs.SetGeneration(ctx, "vs", "/repo/path", "g2", 1))

		setTriggered := false
		cache := upToDateStoragesCache{
			rs:  rs,
			exp: time.Second,
			cacheByVirtualStorage: map[string]expirationCache{
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
			},
		}

		storages := cache.GetSyncedNodes(ctx, nil, "vs", "/repo/path", "g1")
		require.ElementsMatch(t, []string{"g1", "g2"}, storages)
		require.True(t, setTriggered)
	})

	t.Run("repository store returns error", func(t *testing.T) {
		ctx, cancel := testhelper.Context()
		defer cancel()

		logger := testhelper.DiscardTestLogger(t)
		logHook := test.NewLocal(logger)

		rs := datastore.NewMemoryRepositoryStore(map[string][]string{"bad": {"g1", "g2", "g3"}})

		setTriggered := false
		cache := upToDateStoragesCache{
			rs:  rs,
			exp: time.Second,
			cacheByVirtualStorage: map[string]expirationCache{
				"vs": mockExpirationCache{
					SetMethod: func(k string, v interface{}, d time.Duration) {
						setTriggered = true
					},
					GetMethod: func(k string) (interface{}, bool) {
						return nil, false
					},
				},
			},
		}

		storages := cache.GetSyncedNodes(ctx, logger, "vs", "/repo/path", "g1")
		require.ElementsMatch(t, []string{"g1"}, storages)
		require.True(t, setTriggered)
		require.Len(t, logHook.AllEntries(), 1)
		require.Equal(t, "get up to date secondaries", logHook.LastEntry().Message)
		require.Equal(t, logrus.WarnLevel, logHook.LastEntry().Level)
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
