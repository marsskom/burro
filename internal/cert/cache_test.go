package cert

import (
	"crypto/tls"
	"sync"
	"testing"
)

func dummyCert() *tls.Certificate {
	return &tls.Certificate{}
}

func TestNewCertCacheHostKey_NormalizesLowercase(t *testing.T) {
	k := NewCertCacheHostKey("ExAmPlE.COM")

	if k.normalized != "example.com" {
		t.Fatalf("expected example.com, got %s", k.normalized)
	}
}

func TestNewCertCacheHostKey_StripsPort(t *testing.T) {
	k := NewCertCacheHostKey("Example.COM:443")

	if k.normalized != "example.com" {
		t.Fatalf("expected example.com, got %s", k.normalized)
	}
}

func TestNewCertCacheHostKey_PreservesRawValue(t *testing.T) {
	raw := "Example.COM:443"
	k := NewCertCacheHostKey(raw)

	if k.value != raw {
		t.Fatalf("expected raw value preserved, got %s", k.value)
	}
}

func TestCertCache_SetAndGet(t *testing.T) {
	cache := NewCertCache()

	key := NewCertCacheHostKey("Example.com")
	item := &CertCacheItem{Certificate: dummyCert()}

	if err := cache.Set(key, item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := cache.Get(NewCertCacheHostKey("example.com"))
	if !ok {
		t.Fatal("expected item in cache")
	}

	if got != item {
		t.Fatal("expected same pointer")
	}
}

func TestCertCache_Set_DuplicateNormalizedKey(t *testing.T) {
	cache := NewCertCache()

	k1 := NewCertCacheHostKey("Example.com:443")
	k2 := NewCertCacheHostKey("example.com")

	if err := cache.Set(k1, &CertCacheItem{Certificate: dummyCert()}); err != nil {
		t.Fatalf("first set failed: %v", err)
	}

	err := cache.Set(k2, &CertCacheItem{Certificate: dummyCert()})
	if err == nil {
		t.Fatal("expected duplicate error due to normalization")
	}
}

func TestCertCache_Delete(t *testing.T) {
	cache := NewCertCache()

	key := NewCertCacheHostKey("Example.com")

	_ = cache.Set(key, &CertCacheItem{Certificate: dummyCert()})

	cache.Delete(NewCertCacheHostKey("example.com"))

	_, ok := cache.Get(key)
	if ok {
		t.Fatal("expected item to be deleted")
	}
}

func TestCertCache_Clear(t *testing.T) {
	cache := NewCertCache()

	_ = cache.Set(NewCertCacheHostKey("a.com"), &CertCacheItem{Certificate: dummyCert()})
	_ = cache.Set(NewCertCacheHostKey("b.com"), &CertCacheItem{Certificate: dummyCert()})

	cache.Clear()

	if _, ok := cache.Get(NewCertCacheHostKey("a.com")); ok {
		t.Fatal("expected cache to be empty")
	}

	if _, ok := cache.Get(NewCertCacheHostKey("b.com")); ok {
		t.Fatal("expected cache to be empty")
	}
}

func TestCertCache_ConsistencyAcrossRawForms(t *testing.T) {
	cache := NewCertCache()

	_ = cache.Set(NewCertCacheHostKey("Example.Com:443"), &CertCacheItem{Certificate: dummyCert()})

	tests := []string{
		"example.com",
		"Example.COM",
		"EXAMPLE.com:443",
	}

	for _, h := range tests {
		got, ok := cache.Get(NewCertCacheHostKey(h))
		if !ok {
			t.Fatalf("expected hit for %s", h)
		}
		if got == nil {
			t.Fatalf("nil item for %s", h)
		}
	}
}

func TestCertCache_ConcurrentAccess(t *testing.T) {
	cache := NewCertCache()

	key := NewCertCacheHostKey("example.com")

	var wg sync.WaitGroup
	const n = 50

	for range n {
		wg.Go(func() {
			cache.Set(key, &CertCacheItem{Certificate: dummyCert()})
		})
	}

	for range n {
		wg.Go(func() {
			cache.Get(key)
		})
	}

	wg.Wait()

	_, ok := cache.Get(key)
	if !ok {
		t.Fatal("expected item after concurrent operations")
	}
}

func TestCertCache_Concurrent_ReadOnly(t *testing.T) {
	cache := NewCertCache()

	key := NewCertCacheHostKey("eXample.com")

	_ = cache.Set(key, &CertCacheItem{Certificate: dummyCert()})

	var wg sync.WaitGroup
	const n = 100

	for range n {
		wg.Go(func() {
			item, ok := cache.Get(key)
			if !ok || item == nil {
				t.Fatal("unexpected cache miss")
			}
		})
	}

	wg.Wait()
}
