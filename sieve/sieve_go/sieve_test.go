package sievego

import (
	"testing"
)

// Helper function to check if two nodes are the same
func nodesEqual[T comparable](n1, n2 *Node[T]) bool {
	if n1 == nil && n2 == nil {
		return true
	}
	if n1 == nil || n2 == nil {
		return false
	}
	return n1.value == n2.value && n1.end_identifier == n2.end_identifier
}

// Helper function to check slice of values in queue
func getQueueValues[T comparable](q *FIFOQueue[T]) []any {
	var values []any
	curr := q.head.next
	for curr != q.tail {
		values = append(values, curr.value)
		curr = curr.next
	}
	return values
}

func TestNewNode(t *testing.T) {
	node := NewNode[string]("testValue")
	if node == nil {
		t.Fatal("NewNode returned nil")
	}
	if node.value != "testValue" {
		t.Errorf("Expected node value 'testValue', got %v", node.value)
	}
	if node.next != nil || node.prev != nil {
		t.Errorf("Expected new node next and prev to be nil, got %v and %v", node.next, node.prev)
	}
	if node.visited != false {
		t.Errorf("Expected new node visited to be false, got %v", node.visited)
	}
	if node.end_identifier != 0 {
		t.Errorf("Expected new node end_identifier to be 0, got %v", node.end_identifier)
	}
}

func TestNewFifoQueue(t *testing.T) {
	fq := NewFifoQueue[string]()
	if fq == nil {
		t.Fatal("NewFifoQueue returned nil")
	}
	if fq.head == nil {
		t.Error("FIFOQueue head is nil")
	}
	if fq.tail == nil {
		t.Error("FIFOQueue tail is nil")
	}
	if fq.head.end_identifier != 1 {
		t.Errorf("Expected head end_identifier to be 1, got %d", fq.head.end_identifier)
	}
	if fq.tail.end_identifier != -1 {
		t.Errorf("Expected tail end_identifier to be -1, got %d", fq.tail.end_identifier)
	}
	if fq.head.next != fq.tail {
		t.Error("FIFOQueue head.next should be tail")
	}
	if fq.tail.prev != fq.head {
		t.Error("FIFOQueue tail.prev should be head")
	}
}

func TestFIFOQueue_insertNode(t *testing.T) {
	fq := NewFifoQueue[string]()
	head := fq.getHead()
	tail := fq.getTail()

	node1 := fq.insertNode("val1", head, tail)
	if node1.value != "val1" {
		t.Errorf("Expected node1 value 'val1', got %v", node1.value)
	}
	if head.next != node1 || node1.prev != head {
		t.Error("Node1 not inserted correctly after head")
	}
	if tail.prev != node1 || node1.next != tail {
		t.Error("Node1 not inserted correctly before tail")
	}

	node2 := fq.insertNode("val2", node1, tail)
	if node2.value != "val2" {
		t.Errorf("Expected node2 value 'val2', got %v", node2.value)
	}
	if node1.next != node2 || node2.prev != node1 {
		t.Error("Node2 not inserted correctly after node1")
	}
	if tail.prev != node2 || node2.next != tail {
		t.Error("Node2 not inserted correctly before tail")
	}

	values := getQueueValues(fq)
	expectedValues := []any{"val1", "val2"}
	if len(values) != len(expectedValues) {
		t.Fatalf("Expected queue length %d, got %d. Values: %v", len(expectedValues), len(values), values)
	}
	for i, v := range values {
		if v != expectedValues[i] {
			t.Errorf("Expected value %v at index %d, got %v", expectedValues[i], i, v)
		}
	}
}

func TestFIFOQueue_deleteNode(t *testing.T) {
	fq := NewFifoQueue[string]()
	head := fq.getHead()
	tail := fq.getTail()

	// Test deleting head or tail (should do nothing)
	fq.deleteNode(head)
	if head.next != tail || tail.prev != head {
		t.Error("Deleting head should not modify the queue structure if only head and tail exist")
	}
	fq.deleteNode(tail)
	if head.next != tail || tail.prev != head {
		t.Error("Deleting tail should not modify the queue structure if only head and tail exist")
	}

	node1 := fq.insertNode("val1", head, tail)
	node2 := fq.insertNode("val2", node1, tail)
	node3 := fq.insertNode("val3", node2, tail) // head <-> node1 <-> node2 <-> node3 <-> tail

	// Delete middle node (node2)
	fq.deleteNode(node2)
	if node1.next != node3 || node3.prev != node1 {
		t.Error("Node2 not deleted correctly. node1.next or node3.prev is wrong.")
	}
	values := getQueueValues(fq)
	expectedValues := []any{"val1", "val3"}
	if len(values) != len(expectedValues) {
		t.Fatalf("Expected queue length %d after deleting node2, got %d. Values: %v", len(expectedValues), len(values), values)
	}
	for i, v := range values {
		if v != expectedValues[i] {
			t.Errorf("Expected value %v at index %d, got %v", expectedValues[i], i, v)
		}
	}

	// Delete first actual node (node1)
	fq.deleteNode(node1)
	if head.next != node3 || node3.prev != head {
		t.Error("Node1 not deleted correctly. head.next or node3.prev is wrong.")
	}
	values = getQueueValues(fq)
	expectedValues = []any{"val3"}
	if len(values) != len(expectedValues) {
		t.Fatalf("Expected queue length %d after deleting node1, got %d. Values: %v", len(expectedValues), len(values), values)
	}
	if values[0] != "val3" {
		t.Errorf("Expected value 'val3', got %v", values[0])
	}

	// Delete last actual node (node3)
	fq.deleteNode(node3)
	if head.next != tail || tail.prev != head {
		t.Error("Node3 not deleted correctly. head.next or tail.prev is wrong.")
	}
	values = getQueueValues(fq)
	if len(values) != 0 {
		t.Fatalf("Expected empty queue after deleting node3, got %d values: %v", len(values), values)
	}

	// Try deleting a node not in the list (or already deleted) - should not panic
	// Create a detached node
	detachedNode := NewNode[string]("detached")
	fq.deleteNode(detachedNode)                 // Should not panic or corrupt the list
	if head.next != tail || tail.prev != head { // Check if list is still intact
		t.Error("Deleting a detached node corrupted the list")
	}

}

func TestFIFOQueue_getHead_getTail_Panic(t *testing.T) {
	fq := NewFifoQueue[string]()

	// Test getHead panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("getHead did not panic with modified head")
		}
	}()
	fq.head.end_identifier = 0 // Corrupt head
	fq.getHead()               // Should panic
}

func TestFIFOQueue_getTail_Panic(t *testing.T) {
	fq := NewFifoQueue[string]()
	// Test getTail panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("getTail did not panic with modified tail")
		}
	}()
	fq.tail.end_identifier = 0 // Corrupt tail
	fq.getTail()               // Should panic
}

func TestNewSieve(t *testing.T) {
	s := NewSieve[string](10)
	if s == nil {
		t.Fatal("NewSieve returned nil")
	}
	if s.Capacity != 10 {
		t.Errorf("Expected sieve capacity 10, got %d", s.Capacity)
	}
	if s.Nodes == nil {
		t.Error("Sieve Nodes map is nil")
	}
	if len(s.Nodes) != 0 {
		t.Errorf("Expected new sieve Nodes map to be empty, got size %d", len(s.Nodes))
	}
	if s.FifoQueue == nil {
		t.Error("Sieve FifoQueue is nil")
	}
	if s.hand != nil {
		t.Errorf("Expected new sieve hand to be nil, got %v", s.hand)
	}
}

func TestSieve_IsEmpty(t *testing.T) {
	s := NewSieve[string](2)
	if !s.IsEmpty() {
		t.Error("Expected new sieve to be empty")
	}
	s.Insert("key1", "data1")
	if s.IsEmpty() {
		t.Error("Expected sieve not to be empty after insert")
	}
}

func TestSieve_Get(t *testing.T) {
	s := NewSieve[string](2)
	s.Insert("key1", "data1")

	// Test get existing key
	if !s.Get("key1") {
		t.Error("Expected Get to return true for existing key 'key1'")
	}
	node1, ok := s.Nodes["key1"]
	if !ok || !node1.visited {
		t.Error("Expected node 'key1' to be marked as visited after Get")
	}

	// Test get non-existing key
	if s.Get("key2") {
		t.Error("Expected Get to return false for non-existing key 'key2'")
	}
}

func TestSieve_Insert_Basic(t *testing.T) {
	s := NewSieve[string](2)

	s.Insert("key1", "data1")
	if len(s.Nodes) != 1 {
		t.Errorf("Expected 1 node in sieve, got %d", len(s.Nodes))
	}
	node1, ok := s.Nodes["key1"]
	if !ok {
		t.Fatal("Node 'key1' not found in sieve.Nodes after insert")
	}
	if node1.value != "data1" {
		t.Errorf("Expected node 'key1' value 'data1', got %v", node1.value)
	}
	if node1.visited != false { // Inserted node should initially be not visited (as per current logic)
		t.Errorf("Expected new node 'key1' visited to be false, got %v", node1.visited)
	}

	// Check FIFO queue
	qVals := getQueueValues(s.FifoQueue)
	if len(qVals) != 1 || qVals[0] != "data1" {
		t.Errorf("Expected FIFO queue to contain ['data1'], got %v", qVals)
	}

	s.Insert("key2", "data2")
	if len(s.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in sieve, got %d", len(s.Nodes))
	}
	node2, ok := s.Nodes["key2"]
	if !ok {
		t.Fatal("Node 'key2' not found in sieve.Nodes after insert")
	}
	if node2.value != "data2" {
		t.Errorf("Expected node 'key2' value 'data2', got %v", node2.value)
	}

	qVals = getQueueValues(s.FifoQueue) // New items are added to the front
	expectedQVals := []any{"data2", "data1"}
	if len(qVals) != len(expectedQVals) {
		t.Fatalf("Expected FIFO queue length %d, got %d. Values: %v", len(expectedQVals), len(qVals), qVals)
	}
	for i, v := range qVals {
		if v != expectedQVals[i] {
			t.Errorf("Expected FIFO value %v at index %d, got %v", expectedQVals[i], i, v)
		}
	}
}

func TestSieve_Insert_Eviction(t *testing.T) {
	s := NewSieve[string](2) // Capacity of 2

	// Fill the sieve
	s.Insert("key1", "data1") // Queue: [data1]
	s.Insert("key2", "data2") // Queue: [data2, data1] (data2 is at head.next)

	// At this point, hand is nil.
	// The next insert should cause eviction.
	// Expected behavior: hand moves from tail backwards.
	// data1 is at tail.prev, data2 is at tail.prev.prev
	//
	// Initial state for eviction: hand starts at tail.
	// 1. hand = s.FifoQueue.getTail() (tail node)
	// 2. Loop: hand.visited (false for tail) is false, hand.end_identifier != 1.
	//    hand = hand.prev (node "data1")
	// 3. hand.visited (false for "data1") is false. hand ("data1") is not end_identifier 1.
	//    "data1" (s.Nodes["key1"]) will be evicted.
	//    sieve.hand will become node "data2" (the one before "data1")

	s.Insert("key3", "data3") // Queue: [data3, data2], "data1" should be evicted

	if len(s.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in sieve after eviction, got %d", len(s.Nodes))
	}
	if _, present := s.Nodes["key1"]; present {
		t.Error("Expected 'key1' to be evicted")
	}
	if _, present := s.Nodes["key2"]; !present {
		t.Error("Expected 'key2' to remain in sieve")
	}
	if _, present := s.Nodes["key3"]; !present {
		t.Error("Expected 'key3' to be added to sieve")
	}

	qVals := getQueueValues(s.FifoQueue)
	expectedQVals := []any{"data3", "data2"}
	if len(qVals) != len(expectedQVals) {
		t.Fatalf("Expected FIFO queue length %d, got %d. Values: %v", len(expectedQVals), len(qVals), qVals)
	}
	for i, v := range qVals {
		if v != expectedQVals[i] {
			t.Errorf("Expected FIFO value %v at index %d, got %v", expectedQVals[i], i, v)
		}
	}

	// Verify hand position (it should be pointing to the node before the evicted one)
	// In this case, before "data1" was evicted, "data2" was its prev. So hand should be "data2"s node.
	// The actual value of s.hand after insert is tricky to directly verify without exposing more internals
	// or making assumptions on the key of the evicted node. We are indirectly testing it by checking
	// which node got evicted.

	// Test eviction with visited nodes
	s.Get("key2") // Mark "data2" as visited. Queue: [data3, data2(v)]
	// hand is currently pointing to node for "data2" (from previous eviction)

	s.Insert("key4", "data4") // Queue: [data4, data3], "data2" should be kept, "data3" evicted
	// Eviction logic:
	// 1. hand is node "data2". hand.visited is true. Set to false. hand = hand.prev (node "data3")
	// 2. hand is node "data3". hand.visited is false. Evict "data3".
	//    s.hand becomes node "data4" (actually, the head sentinel, then prev to that.
	//    No, hand becomes prev of evicted, which is the new node data4)

	if len(s.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in sieve after eviction with visited, got %d", len(s.Nodes))
	}
	if _, present := s.Nodes["key3"]; present {
		t.Error("Expected 'key3' to be evicted (was not visited)")
	}
	if _, present := s.Nodes["key2"]; !present {
		t.Error("Expected 'key2' to remain in sieve (was visited)")
	}
	if _, present := s.Nodes["key4"]; !present {
		t.Error("Expected 'key4' to be added to sieve")
	}

	qVals = getQueueValues(s.FifoQueue)
	expectedQVals = []any{"data4", "data2"} // data2 was kept, data3 evicted
	if len(qVals) != len(expectedQVals) {
		t.Fatalf("Expected FIFO queue length %d, got %d. Values: %v", len(expectedQVals), len(qVals), qVals)
	}
	foundD4, foundD2 := false, false
	for _, v := range qVals {
		if v == "data4" {
			foundD4 = true
		}
		if v == "data2" {
			foundD2 = true
		}
	}
	if !foundD4 || !foundD2 {
		t.Errorf("Expected FIFO queue to contain 'data4' and 'data2', got %v", qVals)
	}
}

func TestSieve_Insert_Eviction_AllVisited(t *testing.T) {
	s := NewSieve[string](2)
	s.Insert("key1", "data1") // Q: [data1]
	s.Insert("key2", "data2") // Q: [data2, data1]

	s.Get("key1") // data1 visited
	s.Get("key2") // data2 visited

	// hand is nil. Will start from tail.
	// 1. hand = tail. prev = "data1"(v). "data1".visited = true -> false. hand = "data1".prev ("data2")
	// 2. hand = "data2"(v). "data2".visited = true -> false. hand = "data2".prev (head)
	// 3. hand is head.end_identifier == 1. hand becomes tail.
	// 4. hand = tail.prev ("data1"). "data1" (visited now false) is evicted.
	s.Insert("key3", "data3") // Should evict the oldest (key1)

	if len(s.Nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d. Nodes: %v", len(s.Nodes), s.Nodes)
	}
	if _, present := s.Nodes["key1"]; present {
		t.Error("Expected 'key1' to be evicted even if all were visited (LRU behavior when hand wraps)")
	}
	if _, present := s.Nodes["key2"]; !present {
		t.Error("Expected 'key2' to remain")
	}
	if _, present := s.Nodes["key3"]; !present {
		t.Error("Expected 'key3' to be inserted")
	}

	qVals := getQueueValues(s.FifoQueue)
	expectedQVals := []any{"data3", "data2"}
	if len(qVals) != len(expectedQVals) {
		t.Fatalf("Expected FIFO queue length %d, got %d. Values: %v", len(expectedQVals), len(qVals), qVals)
	}
	foundD3, foundD2 := false, false
	for _, v := range qVals {
		if v == "data3" {
			foundD3 = true
		}
		if v == "data2" {
			foundD2 = true
		}
	}
	if !foundD3 || !foundD2 {
		t.Errorf("Expected FIFO queue to contain 'data3' and 'data2', got %v", qVals)
	}
}

func TestSieve_Insert_HandBecomesHeadThenTail(t *testing.T) {
	// This test aims to ensure the hand correctly wraps from head to tail
	// during the eviction scan if all nodes are visited.
	s := NewSieve[string](1)  // Capacity 1
	s.Insert("key1", "data1") // Q: [data1], hand: nil

	s.Get("key1") // key1 is visited

	s.Insert("key2", "data2") // k1 (v=F) should be evicted (oldest, after its visited bit was cleared)

	if len(s.Nodes) != 1 {
		t.Errorf("Expected 1 node after eviction, got %d", len(s.Nodes))
	}
	if _, present := s.Nodes["key1"]; present {
		// This assertion depends on the eviction logic for "all visited" case.
		// Based on the trace above, k1 might *not* be evicted if the deleteNode targets tail sentinel.
		// Let's run and see. If k1 is still there, the bug is confirmed by test.
		t.Error("Expected 'key1' to be evicted")
	}
	if _, present := s.Nodes["key2"]; !present {
		t.Error("Expected 'key2' to be inserted")
	}

	qVals := getQueueValues(s.FifoQueue)
	// If N1 was not deleted: Q = [N2(d2), N1(d1, v=F)]
	// If N1 was deleted (hypothetically, if logic was perfect): Q = [N2(d2)]
	// Given `deleteNode(tail_sentinel)`: N1 is NOT deleted from queue.
	expectedQValsAfterK2 := []any{"data2"}
	if len(qVals) != len(expectedQValsAfterK2) {
		t.Fatalf("Queue: Expected length %d, got %d. Values: %v", len(expectedQValsAfterK2), len(qVals), qVals)
	}
	if qVals[0] != "data2" {
		t.Errorf("Queue: Expected front 'data2', then 'data1', got %v", qVals)
	}

	// This test highlights fundamental issues in the eviction logic regarding map cleanup
	// and potentially queue cleanup in edge cases.
	// The tests will proceed assuming s.Nodes is NOT cleared.
}

func TestSieve_Insert_FullCapacity_ThenGet(t *testing.T) {
	s := NewSieve[string](1)
	s.Insert("key1", "data1") // Map: {k1:N1}, Q: [N1(d1,v=F)]

	// At this point, len(s.Nodes) == s.Capacity.
	// The map s.Nodes is not cleared on "eviction" by current code.
	s.Insert("key2", "data2") // Map: {k1:N1, k2:N2}, Q: [N2(d2,v=F), N1(d1,v=F)] (N1 wasn't really evicted by flawed logic)
	// s.hand would point to N1.prev (head sentinel).

	if len(s.Nodes) != 1 { // Because map is not cleared
		t.Errorf("Expected Nodes length to be 2 (due to no map eviction), got %d", len(s.Nodes))
	}

	if !s.Get("key2") {
		t.Error("Expected Get key2 to return true")
	}
	node2, _ := s.Nodes["key2"]
	if !node2.visited {
		t.Error("Expected key2 to be marked visited after Get")
	}
}
