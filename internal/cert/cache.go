package cert

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
)

type CertCacheHostKey string

func (ck CertCacheHostKey) Normalize() CertCacheHostKey {
	slog.Debug("Normalize cert cache host key", "host", ck)

	host := strings.ToLower(string(ck))

	if h, _, err := net.SplitHostPort(host); err == nil {
		return CertCacheHostKey(h)
	}

	return CertCacheHostKey(host)
}

type CertCacheItem struct {
	Certificate *tls.Certificate
}

type CertCache struct {
	items map[CertCacheHostKey]*CertCacheItem
	mu    sync.RWMutex
}

func NewCertCache() *CertCache {
	return &CertCache{
		items: make(map[CertCacheHostKey]*CertCacheItem),
	}
}

func (c *CertCache) Set(host CertCacheHostKey, item *CertCacheItem) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	host = host.Normalize()

	if _, ok := c.items[host]; ok {
		return fmt.Errorf("cache certificate item with host '%s' already exist", host)
	}

	c.items[host] = item

	slog.Debug("CertCache: set for a host", "host", host)

	return nil
}

func (c *CertCache) Get(host CertCacheHostKey) (*CertCacheItem, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	host = host.Normalize()

	item, ok := c.items[host]

	slog.Debug("CertCache: get for a host", "host", host, "ok", ok)

	return item, ok
}

func (c *CertCache) Delete(host CertCacheHostKey) {
	c.mu.Lock()
	defer c.mu.Unlock()

	host = host.Normalize()

	slog.Debug("CertCache: delete for a host", "host", host)

	delete(c.items, host)
}

func (c *CertCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	slog.Debug("CertCache: clear")

	c.items = make(map[CertCacheHostKey]*CertCacheItem)
}
