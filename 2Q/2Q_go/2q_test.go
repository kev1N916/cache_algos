package qgo

import (
	"testing"
)

// TestNodeOperations tests the basic node operations
func TestNodeOperations(t *testing.T) {
	// Create a simple linked list with nodes
	tail := &Node[string]{end_identifier: -1}
	head := &Node[string]{end_identifier: 1}

	head.next = tail
	tail.prev = head

	// Test creating a new node
	node1 := NewNode("key1", head, tail)

	if node1.key != "key1" {
		t.Errorf("Node key not set correctly, expected 'key1', got '%s'", node1.key)
	}

	if head.next != node1 {
		t.Errorf("Head.next not pointing to new node")
	}

	if tail.prev != node1 {
		t.Errorf("Tail.prev not pointing to new node")
	}

	// Test adding another node
	node2 := NewNode("key2", head, node1)

	if node2.next != node1 {
		t.Errorf("Node2.next not pointing to node1")
	}

	if node1.prev != node2 {
		t.Errorf("Node1.prev not pointing to node2")
	}

	// Test deleting a node
	deleteNode(node1)

	if head.next != node2 || node2.next != tail {
		t.Errorf("Node deletion failed to update pointers correctly")
	}
}

// TestFIFOOperations tests the FIFO queue operations
func TestFIFOOperations(t *testing.T) {
	fifo := NewFIFO[string]()

	// Initialize the nodes map
	fifo.Nodes = make(map[string]*Node[string])

	// Test add operation
	node1 := fifo.add("key1")
	if !fifo.isPresent("key1") {
		t.Errorf("Key 'key1' should be present in FIFO after adding")
	}

	node2 := fifo.add("key2")
	if fifo.Head.next != node2 {
		t.Errorf("New node should be added at the head")
	}

	// Verify the structure: head -> node2 -> node1 -> tail
	if node2.next != node1 || node1.next != fifo.Tail {
		t.Errorf("FIFO structure incorrect after adding nodes")
	}

	// Test eviction (should remove node1 which is closest to tail)
	_, evicted := fifo.evict()
	if !evicted {
		t.Errorf("Eviction should have succeeded")
	}

	// Verify structure after eviction: head -> node2 -> tail
	if fifo.Head.next != node2 || node2.next != fifo.Tail {
		t.Errorf("FIFO structure incorrect after eviction")
	}

	// Evict one more node
	_, evicted = fifo.evict()
	if !evicted {
		t.Errorf("Second eviction should have succeeded")
	}

	// Verify empty structure: head -> tail
	if fifo.Head.next != fifo.Tail {
		t.Errorf("FIFO should be empty after evicting all nodes")
	}

	// Try to evict from empty queue
	_, evicted = fifo.evict()
	if evicted {
		t.Errorf("Eviction from empty FIFO should return false")
	}
}

// TestLRUOperations tests the LRU cache operations
func TestLRUOperations(t *testing.T) {
	lru := NewLRU[string]()

	// Initialize the nodes map
	lru.Nodes = make(map[string]*Node[string])

	// Test add operation
	node1 := lru.add("key1")
	if lru.Nodes["key1"] != node1 {
		t.Errorf("Node not properly stored in nodes map")
	}

	node2 := lru.add("key2")
	if lru.Head.next != node2 {
		t.Errorf("New node should be added at the head")
	}

	// Test access operation (should move node1 to the front)
	lru.access("key1")

	// // Verify structure after access: head -> node1 -> node2 -> tail
	// if lru.Head.next != node1 || node1.next != node2 {
	// 	t.Errorf("LRU structure incorrect after access operation")
	// }
	nextToHead:=lru.getHead().next
	if nextToHead.key != "key1" {
		t.Errorf("LRU structure incorrect after access operation")
	}

	// Test eviction (should remove node2 which is closest to tail)
	keyEvicted, evicted := lru.evict()
	if !evicted {
		t.Errorf("Eviction should have succeeded")
	}
	if !(keyEvicted=="key2") {
		t.Errorf("key2 should be evicted")
	}
}

// TestTwoQBasicOperations tests the basic operations of the TwoQ cache
func TestTwoQBasicOperations(t *testing.T) {
	// Create a new TwoQ cache with capacity 3
	twoQ := NewTwoQ[string](3)

	// Initialize required components
	twoQ.K_In = 1
	twoQ.K_Out = 1
	twoQ.PageBuffer = make(map[string]*Page)
	twoQ.A1in = NewFIFO[string]()
	twoQ.A1in.Nodes = make(map[string]*Node[string])
	twoQ.Am = NewLRU[string]()
	twoQ.Am.Nodes = make(map[string]*Node[string])
	twoQ.A1out = NewFIFO[string]()
	twoQ.A1out.Nodes = make(map[string]*Node[string])

	// Test inserting a new key
	_, present := twoQ.Insert("key1", "value1")
	if present {
		t.Errorf("First insertion should return present=false")
	}

	if len(twoQ.PageBuffer) != 1 {
		t.Errorf("PageBuffer should have 1 entry after first insertion")
	}

	// Insert the same key again
	_, present = twoQ.Insert("key1", "value1-updated")
	if !present {
		t.Errorf("Second insertion of same key should return present=true")
	}

	// Insert until we reach capacity
	_, _ = twoQ.Insert("key2", "value2")
	_, _ = twoQ.Insert("key3", "value3")

	if len(twoQ.PageBuffer) != 3 {
		t.Errorf("PageBuffer should have 3 entries after inserting 3 keys")
	}

	// Insert one more key to trigger eviction
	_, _ = twoQ.Insert("key4", "value4")

	if len(twoQ.PageBuffer) > 3 {
		t.Errorf("PageBuffer should not exceed capacity of 3")
	}
}

// TestTwoQEvictionPolicy tests the eviction policy of the TwoQ cache
func TestTwoQEvictionPolicy(t *testing.T) {
	// Create a new TwoQ cache with capacity 3
	twoQ := NewTwoQ[string](3)

	// Initialize required components
	twoQ.K_In = 1
	twoQ.K_Out = 1
	twoQ.PageBuffer = make(map[string]*Page)
	twoQ.A1in = NewFIFO[string]()
	twoQ.A1in.Nodes = make(map[string]*Node[string])
	twoQ.Am = NewLRU[string]()
	twoQ.Am.Nodes = make(map[string]*Node[string])
	twoQ.A1out = NewFIFO[string]()
	twoQ.A1out.Nodes = make(map[string]*Node[string])

	// Insert keys to fill the cache
	twoQ.Insert("key1", "value1")
	twoQ.Insert("key2", "value2")
	twoQ.Insert("key3", "value3")

	// Access key1 again to promote it to Am
	twoQ.Insert("key1", "value1-again")

	// Insert key4 which should evict key1 (oldest in A1in)
	twoQ.Insert("key4", "value4")

	// Check that key1 is no longer in PageBuffer
	_, present := twoQ.PageBuffer["key1"]
	if present {
		t.Errorf("key1 should have been evicted")
	}

	// But key1 should still be there (in Am)
	present = twoQ.A1out.isPresent("key1")
	if !present {
		t.Errorf("key1 should still be in cache after promotion to Am")
	}
}

// TestTwoQPromotionPolicy tests the promotion policy of the TwoQ cache
func TestTwoQPromotionPolicy(t *testing.T) {
	// Create a new TwoQ cache with capacity 5
	twoQ := NewTwoQ[string](4)

	// Initialize required components
	twoQ.K_In = 2
	twoQ.K_Out = 2
	twoQ.PageBuffer = make(map[string]*Page)
	twoQ.A1in = NewFIFO[string]()
	twoQ.A1in.Nodes = make(map[string]*Node[string])
	twoQ.Am = NewLRU[string]()
	twoQ.Am.Nodes = make(map[string]*Node[string])
	twoQ.A1out = NewFIFO[string]()
	twoQ.A1out.Nodes = make(map[string]*Node[string])

	// Insert some keys
	twoQ.Insert("key1", "value1")
	twoQ.Insert("key2", "value2")
	twoQ.Insert("key3", "value3")

	// Force eviction of key1 and key2
	twoQ.Insert("key4", "value4")
	twoQ.Insert("key5", "value5")
	twoQ.Insert("key6", "value6")


	// Now key2 and key1 should be in A1out
	// Access key2 again which should promote it to Am
	twoQ.Insert("key2", "value2-again")

	// Check that key2 is in PageBuffer with queueType "A_M"
	page, present := twoQ.PageBuffer["key2"]
	if !present {
		t.Errorf("key2 should be present after being re-accessed")
	}

	if page.queueType != "A_M" {
		t.Errorf("key2 should be in A_M after being re-accessed, got: %s", page.queueType)
	}

	// key1 should still be in A1out and not in PageBuffer
	_, present = twoQ.PageBuffer["key1"]
	if present {
		t.Errorf("key3 should not be in PageBuffer")
	}
}

// TestFIFOEmpty tests the behavior of FIFO when empty
func TestFIFOEmpty(t *testing.T) {
	fifo := NewFIFO[string]()
	fifo.Nodes = make(map[string]*Node[string])

	// Check empty status
	if fifo.Head.next != fifo.Tail {
		t.Errorf("Newly created FIFO should be empty")
	}

	// Try to evict from empty FIFO
	_, evicted := fifo.evict()
	if evicted {
		t.Errorf("Eviction from empty FIFO should return false")
	}
}

// TestLRUEmpty tests the behavior of LRU when empty
func TestLRUEmpty(t *testing.T) {
	lru := NewLRU[string]()
	lru.Nodes = make(map[string]*Node[string])

	// Check empty status
	if lru.Head.next != lru.Tail {
		t.Errorf("Newly created LRU should be empty")
	}

	// Try to evict from empty LRU
	_, evicted := lru.evict()
	if evicted {
		t.Errorf("Eviction from empty LRU should return false")
	}
}
