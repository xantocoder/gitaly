package datastore

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// ErrVirtualStorageNotExist indicates that the requested virtual storage can't be found or not configured.
var ErrVirtualStorageNotExist = errors.New("virtual storage does not exist")

// UpToDateStoragesCache is a storage provider that caches up to date storages by repository.
// Each virtual storage has it's own cache that invalidates entries based on notifications or if an expiration
// passes.
type UpToDateStoragesCache struct {
	// callbackLogger should be used only inside of the methods used as callbacks.
	callbackLogger        logrus.FieldLogger
	rs                    RepositoryStore
	mtx                   sync.Mutex
	cacheByVirtualStorage map[string]expirationCache
	expiration            time.Duration
	errorsTotal           *prometheus.CounterVec
	cacheAccessTotal      *prometheus.CounterVec
}

// NewUpToDateStoragesCache returns a storage provider that uses caching.
func NewUpToDateStoragesCache(logger logrus.FieldLogger, rs RepositoryStore, expiration time.Duration, virtualStorages []string) *UpToDateStoragesCache {
	sc := UpToDateStoragesCache{
		callbackLogger:        logger.WithField("component", "up_to_date_storages_cache"),
		rs:                    rs,
		cacheByVirtualStorage: make(map[string]expirationCache, len(virtualStorages)),
		expiration:            expiration,
		errorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "gitaly_praefect_uptodate_storages_errors_total",
				Help: "Total number of errors raised during defining up to date storages for reads distribution",
			},
			[]string{"type"},
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
		virtualStorage := virtualStorage
		cache := gocache.New(expiration, 30*time.Second)
		cache.OnEvicted(func(string, interface{}) { sc.cacheAccessTotal.WithLabelValues(virtualStorage, "evict").Inc() })
		sc.cacheByVirtualStorage[virtualStorage] = cache
	}

	return &sc
}

// expirationCache is an abstraction over github.com/patrickmn/go-cache for testing purposes.
// Please refer to gocache.Cache for the documentation.
type expirationCache interface {
	Set(k string, v interface{}, d time.Duration)
	Get(k string) (interface{}, bool)
	Delete(k string)
}

func (c *UpToDateStoragesCache) GetSyncedNodes(ctx context.Context, logger logrus.FieldLogger, virtualStorageName, repoPath, primaryStorage string) []string {
	storages, found := c.retrieveFromCache(virtualStorageName, repoPath)
	if found {
		c.cacheAccessTotal.WithLabelValues(virtualStorageName, "hit").Inc()
		return storages
	}

	c.mtx.Lock()
	defer c.mtx.Unlock()

	// another concurrent request may populate the cache with data already
	storages, found = c.retrieveFromCache(virtualStorageName, repoPath)
	if found {
		c.cacheAccessTotal.WithLabelValues(virtualStorageName, "hit").Inc()
		return storages
	}

	c.cacheAccessTotal.WithLabelValues(virtualStorageName, "miss").Inc()

	upToDateStorages, err := c.rs.GetConsistentSecondaries(ctx, virtualStorageName, repoPath, primaryStorage)
	if err != nil {
		c.errorsTotal.WithLabelValues("sql_retrieve").Inc()
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

func (c *UpToDateStoragesCache) retrieveFromCache(virtualStorageName, repoPath string) ([]string, bool) {
	val, found := c.cacheByVirtualStorage[virtualStorageName].Get(repoPath)
	if !found {
		return nil, false
	}

	if val == nil {
		return nil, true
	}

	return val.([]string), true
}

func (c *UpToDateStoragesCache) storeToCache(virtualStorageName, repoPath string, storages []string) {
	c.cacheByVirtualStorage[virtualStorageName].Set(repoPath, storages, c.expiration)
}

func (c *UpToDateStoragesCache) removeFromCache(virtualStorageName, repoPath string) error {
	cache, found := c.cacheByVirtualStorage[virtualStorageName]
	if !found {
		return ErrVirtualStorageNotExist
	}

	cache.Delete(repoPath)

	return nil
}

type notificationEntry struct {
	VirtualStorage string `json:"virtual_storage"`
	RelativePath   string `json:"relative_path"`
}

type changeNotification struct {
	Old []notificationEntry `json:"old"`
	New []notificationEntry `json:"new"`
}

func (c *UpToDateStoragesCache) Notification(payload string) {
	var change changeNotification
	if err := json.NewDecoder(strings.NewReader(payload)).Decode(&change); err != nil {
		c.errorsTotal.WithLabelValues("notification_decode").Inc()
		c.callbackLogger.WithError(err).Error("received payload can't be processed")
		return
	}

	for _, notificationEntries := range [][]notificationEntry{change.Old, change.New} {
		for _, entry := range notificationEntries {
			if err := c.removeFromCache(entry.VirtualStorage, entry.RelativePath); err != nil {
				c.errorsTotal.WithLabelValues("cache_entry_removal").Inc()
				c.callbackLogger.WithError(err).WithField("entry", entry).Error("can't be removed from the cache")
				continue
			}

			c.cacheAccessTotal.WithLabelValues(entry.VirtualStorage, "evict").Inc()
		}
	}
}

func (c *UpToDateStoragesCache) Disconnect() {
	newCache := make(map[string]expirationCache, len(c.cacheByVirtualStorage))
	for vs := range c.cacheByVirtualStorage {
		newCache[vs] = gocache.New(c.expiration, 30*time.Second)
	}

	c.mtx.Lock()
	c.cacheByVirtualStorage = newCache
	c.mtx.Unlock()
}

func (c *UpToDateStoragesCache) Connected() {}

func (c *UpToDateStoragesCache) Describe(descs chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, descs)
}

func (c *UpToDateStoragesCache) Collect(collector chan<- prometheus.Metric) {
	c.errorsTotal.Collect(collector)
	c.cacheAccessTotal.Collect(collector)
}

// DumbStorageProvider is a dumb mechanism to get list of synchronised nodes of the virtual storage.
// It doesn't do any caching or other kinds of improvements and only uses 'datastore.RepositoryStore'
// to get the result.
// It should be used in the tests or in case there is no connection to PostgreSQL database.
type DumbStorageProvider struct {
	RepositoryStore
}

func (sp DumbStorageProvider) GetSyncedNodes(ctx context.Context, logger logrus.FieldLogger, virtualStorageName, repoPath, primaryStorage string) []string {
	upToDateStorages, err := sp.GetConsistentSecondaries(ctx, virtualStorageName, repoPath, primaryStorage)
	if err != nil {
		// this is recoverable error - we can proceed with primary node
		logger.WithError(err).Warn("dumb impl: get up to date secondaries")
	}

	if len(upToDateStorages) == 0 {
		if upToDateStorages == nil {
			upToDateStorages = make(map[string]struct{}, 1)
		}
	}

	// primary should be considered as all other storages for serving read operations
	upToDateStorages[primaryStorage] = struct{}{}
	var storages []string
	for upToDateStorage := range upToDateStorages {
		storages = append(storages, upToDateStorage)
	}

	return storages
}
