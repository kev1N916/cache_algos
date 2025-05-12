package qgo

type TwoQ[T comparable] struct {
	K_In       int
	K_Out      int
	PageBuffer map[T]*Page
	Capacity   int // capacity of page slots
	A1in       *FIFO[T]
	Am         *LRU[T]
	A1out      *FIFO[T]
}

type Page struct {
	data      any
	queueType string
}

type Node[T comparable] struct {
	key            T
	end_identifier int
	prev           *Node[T]
	next           *Node[T]
}

func DeleteNode[T comparable](node *Node[T]) {
	next := node.next
	prev := node.prev

	prev.next = next
	next.prev = prev
}

func NewNode[T comparable](key T, prev, next *Node[T]) *Node[T] {
	node := &Node[T]{
		key:            key,
		end_identifier: 0,
		prev:           prev,
		next:           next,
	}

	prev.next = node
	next.prev = node

	return node
}

type FIFO[T comparable] struct {
	Nodes map[T]*Node[T]
	Tail  *Node[T]
	Head  *Node[T]
}

func NewFIFO[T comparable]() *FIFO[T] {
	tail := &Node[T]{end_identifier: -1}
	head := &Node[T]{end_identifier: 1}

	head.next = tail
	tail.prev = head

	return &FIFO[T]{
		Tail: tail,
		Head: head,
	}
}

func (fifo *FIFO[T]) isPresent(key T) bool {
	_, present := fifo.Nodes[key]
	return present
}

func (fifo *FIFO[T]) add(key T) *Node[T] {
	head := fifo.Head

	if head.end_identifier != 1 {
		panic("this has to be head")
	}

	next_to_head := head.next

	newNode := NewNode(key, head, next_to_head)

	fifo.Nodes[key] = newNode
	return newNode
}

func (fifo *FIFO[T]) evict() (key T, evicted bool) {
	var defaultValue T
	tail := fifo.Tail
	if tail.end_identifier != -1 {
		panic("tail is not initialized properly")
	}

	prev := tail.prev

	if prev.end_identifier == 1 {
		return defaultValue, false
	}

	DeleteNode(prev)
	delete(fifo.Nodes, prev.key)
	return defaultValue, true

}

type LRU[T comparable] struct {
	Nodes map[T]*Node[T]
	Head  *Node[T]
	Tail  *Node[T]
}

// func (lru *LRU[T])
func NewLRU[T comparable]() *LRU[T] {
	tail := &Node[T]{end_identifier: -1}
	head := &Node[T]{end_identifier: 1}

	head.next = tail
	tail.prev = head

	return &LRU[T]{
		Tail: tail,
		Head: head,
	}
}

func (lru *LRU[T]) add(key T) *Node[T] {
	head := lru.Head

	if head.end_identifier != 1 {
		panic("this has to be head")
	}

	next_to_head := head.next

	newNode := NewNode(key, head, next_to_head)

	lru.Nodes[key] = newNode
	return newNode
}

func (lru *LRU[T]) access(key T) {
	node := lru.Nodes[key]
	DeleteNode(node)
	head := lru.getHead()

	node.next = head.next
	head.next = node
}

func (lru *LRU[T]) evict() (T, bool) {
	var defaultValue T
	tail := lru.Tail
	if tail.end_identifier != -1 {
		panic("tail is not initialized properly")
	}

	prev := tail.prev

	if prev.end_identifier == 1 {
		return defaultValue, false
	}

	DeleteNode(prev)
	delete(lru.Nodes, prev.key)
	return defaultValue, true

}

func (lru *LRU[T]) getHead() *Node[T] {
	head := lru.Head
	if head.end_identifier != 1 {
		panic("how")
	}
	return head
}

func NewTwoQ[T comparable](capacity int) *TwoQ[T] {

	return &TwoQ[T]{
		Capacity: capacity,
	}
}

func (twoQ *TwoQ[T]) reclaimFor() {

	if len(twoQ.PageBuffer) < twoQ.Capacity {
		return
	}
	if len(twoQ.A1in.Nodes) >= twoQ.K_In {
		key, evicted := twoQ.A1in.evict()
		if !evicted {
			panic("why cant we evict")
		}

		delete(twoQ.PageBuffer, key)
		twoQ.A1out.add(key)

		if len(twoQ.A1out.Nodes) >= twoQ.K_Out {
			_, evicted := twoQ.A1out.evict()
			if !evicted {
				panic("why cant we evict")
			}
		}
		return
	}

	key, evicted := twoQ.Am.evict()
	if !evicted {
		panic("why cant we evict")
	}
	delete(twoQ.PageBuffer, key)

}
func (twoQ *TwoQ[T]) Insert(key T, value any) (any, bool) {

	page, present := twoQ.PageBuffer[key]

	if !present {
		twoQ.reclaimFor()

		twoQ.PageBuffer[key] = &Page{
			data:      value,
			queueType: "A1_In",
		}

		return page.data, false
	}

	if page.queueType == "A1_In" {
		return page.data, true
	}

	if page.queueType == "A_M" {
		twoQ.Am.access(key)
	}

	if twoQ.A1out.isPresent(key) {
		twoQ.reclaimFor()
		twoQ.Am.add(key)
		twoQ.PageBuffer[key] = &Page{
			data:      value,
			queueType: "A_M",
		}
	}

	return page.data, true

}
