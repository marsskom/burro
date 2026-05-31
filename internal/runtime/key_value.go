package runtime

import (
	"fmt"
	"strings"
	"sync"
)

type KeyValue struct {
	storage map[string][]byte

	mu sync.RWMutex
}

func NewKeyValue() *KeyValue {
	return &KeyValue{
		storage: make(map[string][]byte),
	}
}

func (kv *KeyValue) Get(key string) ([]byte, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	val, ok := kv.storage[key]
	if !ok {
		return []byte{}, fmt.Errorf("kv: key '%s' not found not found", key)
	}

	return val, nil
}

func (kv *KeyValue) Set(key string, value []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if _, ok := kv.storage[key]; ok {
		return fmt.Errorf("kv: key '%s' already exists", key)
	}

	kv.storage[key] = value

	return nil
}

func (kv *KeyValue) Delete(key string) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	delete(kv.storage, key)

	return nil
}

func (kv *KeyValue) List(prefix string) (map[string][]byte, error) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	if prefix == "" {
		return kv.storage, nil
	}

	out := make(map[string][]byte)
	for k, v := range kv.storage {
		if strings.HasPrefix(k, prefix) {
			out[k] = v
		}
	}

	return out, nil
}
