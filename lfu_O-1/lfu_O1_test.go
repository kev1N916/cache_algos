package lfuo1

import (
	"testing"
)

// TestNewLfuCache tests the creation of a new LFU cache
func TestNewLfuCache(t *testing.T) {
	cache := NewLfuCache[string]()
	if cache == nil {
		t.Fatal("Expected non-nil cache")
	}
	if cache.bykey == nil {
		t.Fatal("Expected non-nil bykey map")
	}
	if cache.freq_Head == nil {
		t.Fatal("Expected non-nil freq_Head")
	}
}

// TestInsert tests inserting items into the cache
func TestInsert(t *testing.T) {
	cache := NewLfuCache[string]()
	cache.size = 10 // Set capacity

	// Insert a value
	cache.Insert("key1", "value1")
	
	// Verify it was added
	if len(cache.bykey) != 1 {
		t.Errorf("Expected cache to have 1 item, got %d", len(cache.bykey))
	}
	
	// Check frequency node creation
	if cache.freq_Head.next == nil {
		t.Fatal("Expected freq node to be created")
	}
	if cache.freq_Head.next.value != 1 {
		t.Errorf("Expected frequency 1, got %d", cache.freq_Head.next.value)
	}
	
	// Insert another value
	cache.Insert("key2", "value2")
	
	// Verify it was added
	if len(cache.bykey) != 2 {
		t.Errorf("Expected cache to have 2 items, got %d", len(cache.bykey))
	}
}

// TestInsertPanic tests that inserting a duplicate key panics
func TestInsertPanic(t *testing.T) {
	cache := NewLfuCache[string]()
	cache.size = 10
	
	// Insert a value
	cache.Insert("key1", "value1")
	
	// Insert duplicate key should panic
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic when inserting duplicate key")
		}
	}()
	
	cache.Insert("key1", "value2")
}

// TestAccess tests accessing items changes their frequency
func TestAccess(t *testing.T) {
	cache := NewLfuCache[string]()
	cache.size = 10
	
	// Insert values
	cache.Insert("key1", "value1")
	cache.Insert("key2", "value2")
	
	// Access a value
	val := cache.Access("key1")
	
	// Verify returned value
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}
	
	// Verify frequency was updated
	item := cache.bykey["key1"]
	if item.parent.value != 2 {
		t.Errorf("Expected frequency 2, got %d", item.parent.value)
	}
	
	// Frequency 1 node should still exist with key2
	if cache.freq_Head.next.value != 1 {
		t.Errorf("Expected frequency 1 node to exist, got %d", cache.freq_Head.next.value)
	}
	
	// Access key1 again
	cache.Access("key1")
	
	// Verify frequency was updated to 3
	if cache.bykey["key1"].parent.value != 3 {
		t.Errorf("Expected frequency 3, got %d", cache.bykey["key1"].parent.value)
	}
}

// TestAccessPanic tests that accessing a non-existent key panics
func TestAccessPanic(t *testing.T) {
	cache := NewLfuCache[string]()
	cache.size = 10
	
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic when accessing non-existent key")
		}
	}()
	
	cache.Access("nonexistent")
}

// TestEvict tests the eviction process
func TestEvict(t *testing.T) {
	cache := NewLfuCache[string]()
	cache.size = 2
	
	// Insert values
	cache.Insert("key1", "value1")
	cache.Insert("key2", "value2")
	
	// Access key2 to increase its frequency
	cache.Access("key2")
	
	// Insert a third value which should trigger eviction of key1 (lowest frequency)
	cache.Insert("key3", "value3")
	
	// key1 should be evicted
	if _, exists := cache.bykey["key1"]; exists {
		t.Error("Expected key1 to be evicted")
	}
	
	// key2 and key3 should still exist
	if _, exists := cache.bykey["key2"]; !exists {
		t.Error("Expected key2 to remain in cache")
	}
	if _, exists := cache.bykey["key3"]; !exists {
		t.Error("Expected key3 to be in cache")
	}
}

// TestEvictExplicit tests the manual eviction method
func TestEvictExplicit(t *testing.T) {
	cache := NewLfuCache[string]()
	cache.size = 10
	
	// Insert values
	cache.Insert("key1", "value1")
	cache.Insert("key2", "value2")
	
	// Access key2 to increase its frequency
	cache.Access("key2")
	
	// Manually evict - should remove key1 as it has lowest frequency
	key, value := cache.Evict()
	
	// Verify key1 was evicted
	if key != "key1" {
		t.Errorf("Expected key1 to be evicted, got %v", key)
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}
	
	// Verify key1 is no longer in the cache
	if _, exists := cache.bykey["key1"]; exists {
		t.Error("Expected key1 to be removed from cache")
	}
	
	// key2 should still exist
	if _, exists := cache.bykey["key2"]; !exists {
		t.Error("Expected key2 to remain in cache")
	}
}

// TestEvictPanic tests that evicting from an empty cache panics
func TestEvictPanic(t *testing.T) {
	cache := NewLfuCache[string]()
	
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic when evicting from empty cache")
		}
	}()
	
	cache.Evict()
}

// TestVariousTypes tests the cache with different types of keys
func TestVariousTypes(t *testing.T) {
	// Test with int keys
	intCache := NewLfuCache[int]()
	intCache.size = 10
	intCache.Insert(1, "value1")
	val := intCache.Access(1)
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}
	
	// Test with struct keys
	type CustomKey struct {
		ID int
		Name string
	}
	structCache := NewLfuCache[CustomKey]()
	structCache.size = 10
	key := CustomKey{1, "test"}
	structCache.Insert(key, "value1")
	val = structCache.Access(key)
	if val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}
}

// TestAutoEviction tests that items are automatically evicted when the cache reaches capacity
func TestAutoEviction(t *testing.T) {
	cache := NewLfuCache[string]()
	cache.size = 2
	
	// Insert values up to capacity
	cache.Insert("key1", "value1")
	cache.Insert("key2", "value2")
	
	// Access key2 to increase its frequency
	cache.Access("key2")
	
	// This should trigger automatic eviction of key1 (lowest frequency)
	cache.Insert("key3", "value3")
	
	// Verify key1 was evicted
	if _, exists := cache.bykey["key1"]; exists {
		t.Error("Expected key1 to be evicted")
	}
	
	// key2 and key3 should still exist
	if _, exists := cache.bykey["key2"]; !exists {
		t.Error("Expected key2 to remain in cache")
	}
	if _, exists := cache.bykey["key3"]; !exists {
		t.Error("Expected key3 to be in cache")
	}
}

// TestFrequencyNodeCreationAndDeletion tests the creation and deletion of frequency nodes
func TestFrequencyNodeCreationAndDeletion(t *testing.T) {
	cache := NewLfuCache[string]()
	cache.size = 10
	
	// Insert a value
	cache.Insert("key1", "value1")
	
	// There should be a frequency 1 node
	if cache.freq_Head.next == nil || cache.freq_Head.next.value != 1 {
		t.Fatal("Expected frequency 1 node")
	}
	
	// Access to create frequency 2 node
	cache.Access("key1")
	
	// Verify frequency nodes
	freq1 := cache.freq_Head.next
	if freq1.value != 2 {
		t.Fatal("Expected frequency 1 node to still exist")
	}
	
	// Evict the only item
	cache.Evict()
	
	// All frequency nodes should be gone except head
	if cache.freq_Head.next != nil {
		t.Fatal("Expected all frequency nodes to be deleted")
	}
}