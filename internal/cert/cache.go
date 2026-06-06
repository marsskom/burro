package cert

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"

	"gitlab.com/marsskom/burro/internal/logger"
)

type CertCacheHostKey struct {
	value      string
	normalized string
}

func NewCertCacheHostKey(host string) CertCacheHostKey {
	n := strings.ToLower(host)

	if h, _, err := net.SplitHostPort(n); err == nil {
		n = h
	}

	return CertCacheHostKey{
		value:      host,
		normalized: n,
	}
}

type CertCacheItem struct {
	Certificate *tls.Certificate
}

type CertCache struct {
	items map[string]*CertCacheItem
	mu    sync.RWMutex
}

func NewCertCache() *CertCache {
	return &CertCache{
		items: make(map[string]*CertCacheItem),
	}
}

func (c *CertCache) Set(host CertCacheHostKey, item *CertCacheItem) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.items[host.normalized]; ok {
		return fmt.Errorf("cache certificate item with host '%s' already exist", host.normalized)
	}

	c.items[host.normalized] = item

	logger.Debug("certificate was added to cache for a host", "host", host.normalized)

	return nil
}

func (c *CertCache) Get(host CertCacheHostKey) (*CertCacheItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[host.normalized]

	logger.Debug("get from certificates cache for a host", "host", host.normalized, "ok", ok)

	return item, ok
}

func (c *CertCache) Delete(host CertCacheHostKey) {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.Debug("delete from certificates cache for a host", "host", host.normalized)

	delete(c.items, host.normalized)
}

func (c *CertCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	logger.Debug("clear certificates cache")

	c.items = make(map[string]*CertCacheItem)
}
