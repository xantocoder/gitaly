package nodes

import (
	"context"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gitlab.com/gitlab-org/gitaly/internal/praefect/datastore"
)

type upToDateStoragesCache struct {
	rs                    datastore.RepositoryStore
	mtx                   sync.Mutex
	cacheByVirtualStorage map[string]expirationCache
	expiration            time.Duration
	errorsTotal           prometheus.Counter
	cacheAccessTotal      *prometheus.CounterVec
}

// newUpToDateStoragesCache returns a storage provider that caches up to date storages by repository.
// Each virtual storage has it's own cache that uses `expiration` duration before cache entry would be marked
// as outdated and the value would be re-populated to the cache.
func newUpToDateStoragesCache(rs datastore.RepositoryStore, expiration time.Duration, virtualStorages []string) *upToDateStoragesCache {
	sc := upToDateStoragesCache{
		rs:                    rs,
		cacheByVirtualStorage: make(map[string]expirationCache, len(virtualStorages)),
		expiration:            expiration,
		errorsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "gitaly_praefect_uptodate_storages_errors_total",
				Help: "Total number of errors raised during defining up to date storages for reads distribution",
			},
		),
		cacheAccessTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gitaly_praefect_uptodate_storages_cache_access_total",
				Help: "Total number of cache access operations during defining of up to date storages for reads distribution (per virtual storage)",
			},
			[]string{"virtual_storage", "type"},
		),
	}

	for _, virtualStorage := range virtualStorages {
		sc.cacheByVirtualStorage[virtualStorage] = gocache.New(expiration, 30*time.Second)
	}

	return &sc
}

// is a protection over implementing of the 3-rd party dependency
var _ expirationCache = &gocache.Cache{}

// expirationCache is an abstraction over github.com/patrickmn/go-cache for testing purposes.
// Please refer to gocache.Cache for the documentation.
type expirationCache interface {
	Set(k string, v interface{}, d time.Duration)
	Get(k string) (interface{}, bool)
}

func (c *upToDateStoragesCache) GetSyncedNodes(ctx context.Context, logger logrus.FieldLogger, virtualStorageName, repoPath, primaryStorage string) []string {
	storages, found := c.retrieveFromCache(virtualStorageName, repoPath)
	if found {
		c.cacheAccessTotal.WithLabelValues(virtualStorageName, "hit").Inc()
		return storages
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.cacheAccessTotal.WithLabelValues(virtualStorageName, "miss").Inc()

	// another concurrent request may populate the cache with data already
	storages, found = c.retrieveFromCache(virtualStorageName, repoPath)
	if found {
		return storages
	}

	upToDateStorages, err := c.rs.GetConsistentSecondaries(ctx, virtualStorageName, repoPath, primaryStorage)
	if err != nil {
		c.errorsTotal.Inc()
		// this is recoverable error - we can proceed with primary node
		logger.WithError(err).Warn("get up to date secondaries")
	}

	if len(upToDateStorages) == 0 {
		if upToDateStorages == nil {
			upToDateStorages = make(map[string]struct{}, 1)
		}
	}

	// primary should be considered as all other storages for serving read operations
	upToDateStorages[primaryStorage] = struct{}{}
	for upToDateStorage := range upToDateStorages {
		storages = append(storages, upToDateStorage)
	}

	c.storeToCache(virtualStorageName, repoPath, storages)
	c.cacheAccessTotal.WithLabelValues(virtualStorageName, "populate").Inc()

	return storages
}

func (c *upToDateStoragesCache) retrieveFromCache(virtualStorageName, repoPath string) ([]string, bool) {
	val, found := c.cacheByVirtualStorage[virtualStorageName].Get(repoPath)
	if !found {
		return nil, false
	}

	if val == nil {
		return nil, true
	}

	return val.([]string), true
}

func (c *upToDateStoragesCache) storeToCache(virtualStorageName, repoPath string, storages []string) {
	c.cacheByVirtualStorage[virtualStorageName].Set(repoPath, storages, c.expiration)
}

func (c *upToDateStoragesCache) Describe(descs chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, descs)
}

func (c *upToDateStoragesCache) Collect(collector chan<- prometheus.Metric) {
	c.errorsTotal.Collect(collector)
	c.cacheAccessTotal.Collect(collector)
}
