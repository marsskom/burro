package runtime

import (
	"fmt"
	"sync"
	"testing"
)

func TestKeyValue_SetGet(t *testing.T) {
	kv := NewKeyValue()

	err := kv.Set("a", []byte("1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := kv.Get("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if string(val) != "1" {
		t.Fatalf("expected 1, got %s", val)
	}
}

func TestKeyValue_GetMissing(t *testing.T) {
	kv := NewKeyValue()

	_, err := kv.Get("missing")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestKeyValue_SetDuplicate(t *testing.T) {
	kv := NewKeyValue()

	_ = kv.Set("a", []byte("1"))

	err := kv.Set("a", []byte("2"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := kv.Get("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "2" {
		t.Fatalf("expected 2, got %s", val)
	}
}

func TestKeyValue_Delete(t *testing.T) {
	kv := NewKeyValue()

	_ = kv.Set("a", []byte("1"))
	_ = kv.Delete("a")

	_, err := kv.Get("a")
	if err == nil {
		t.Fatal("expected missing key after delete")
	}
}

func TestKeyValue_List_All(t *testing.T) {
	kv := NewKeyValue()

	_ = kv.Set("a1", []byte("1"))
	_ = kv.Set("b1", []byte("2"))

	out, err := kv.List("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(out))
	}
}

func TestKeyValue_List_Prefix(t *testing.T) {
	kv := NewKeyValue()

	_ = kv.Set("user:1", []byte("a"))
	_ = kv.Set("user:2", []byte("b"))
	_ = kv.Set("system:1", []byte("c"))

	out, err := kv.List("user:")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(out) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(out))
	}
}

func TestKeyValue_ConcurrentAccess(t *testing.T) {
	kv := NewKeyValue()

	var wg sync.WaitGroup

	// Writes.
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			key := fmt.Sprintf("k%d", i)
			_ = kv.Set(key, []byte("v"))
		}(i)
	}

	// Readers.
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			key := fmt.Sprintf("k%d", i)
			kv.Get(key)
		}(i)
	}

	wg.Wait()
}

func TestKeyValue_List_MapMutationRisk(t *testing.T) {
	kv := NewKeyValue()

	_ = kv.Set("a", []byte("1"))

	out, _ := kv.List("")

	// Mutate returned map.
	out["a"] = []byte("CORRUPTED")

	val, _ := kv.Get("a")

	if string(val) != "1" {
		t.Fatal("internal state was corrupted via List()")
	}
}
