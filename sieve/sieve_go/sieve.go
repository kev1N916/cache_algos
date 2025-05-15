package sievego

type Node[T comparable] struct {
	key            T
	end_identifier int
	next           *Node[T]
	prev           *Node[T]
	value          any
	visited        bool
}

func NewNode[T comparable](value any) *Node[T] {

	return &Node[T]{
		value:   value,
		visited: false,
	}

}

type Sieve[T comparable] struct {
	hand      *Node[T]
	Capacity  int
	Nodes     map[T]*Node[T]
	FifoQueue *FIFOQueue[T]
}

func NewSieve[T comparable](cap int) *Sieve[T] {

	if (cap<=0){
		panic("capacity has to be greater than 0")
	}
	return &Sieve[T]{
		Capacity:  cap,
		Nodes:     make(map[T]*Node[T]),
		FifoQueue: NewFifoQueue[T](),
		hand:      nil,
	}
}

type FIFOQueue[T comparable] struct {
	head *Node[T]
	tail *Node[T]
}

func NewFifoQueue[T comparable]() *FIFOQueue[T] {

	head := &Node[T]{
		visited:        false,
		end_identifier: 1,
	}

	tail := &Node[T]{
		visited:        false,
		end_identifier: -1,
	}
	head.next = tail
	tail.prev = head

	return &FIFOQueue[T]{
		head: head,
		tail: tail,
	}

}

func (sieve *Sieve[T]) IsEmpty() bool {
	return len(sieve.Nodes) == 0
}

func (sieve *Sieve[T]) getHand() *Node[T] {
	return sieve.hand
}
func (sieve *Sieve[T]) Get(key T) bool {

	node, present := sieve.Nodes[key]
	if !present {
		return false
	}

	node.visited = true
	return true

}

func (fifoQueue *FIFOQueue[T]) getHead() *Node[T] {
	head := fifoQueue.head
	if head.end_identifier != 1 {
		panic("head is not initialized properly")
	}

	return head
}

func (fifoQueue *FIFOQueue[T]) getTail() *Node[T] {
	tail := fifoQueue.tail
	if tail.end_identifier != -1 {
		panic("tail is not initialized properly")
	}
	return tail
}

func (fifoQueue *FIFOQueue[T]) deleteNode(node *Node[T]) {
	if node.end_identifier == -1 || node.end_identifier == 1 {
		return
	}

	if node.next == nil || node.prev == nil {
		return
	}

	prev := node.prev
	next := node.next

	prev.next = next
	next.prev = prev
}

func (fifoQueue *FIFOQueue[T]) insertNode(value any, prev, next *Node[T]) *Node[T] {
	curr := NewNode[T](value)

	curr.next = next
	next.prev = curr

	prev.next = curr
	curr.prev = prev

	return curr
}
func (sieve *Sieve[T]) Insert(key T, data any) {
	if len(sieve.Nodes) == sieve.Capacity {
		hand := sieve.getHand()
		if hand == nil || hand.end_identifier == 1 {
			hand = sieve.FifoQueue.getTail()
		}

		hand = hand.prev

		for hand.visited  {
			hand.visited = false
			hand = hand.prev

			if hand.end_identifier == 1 {
				hand = sieve.FifoQueue.getTail()
				hand=hand.prev
			}
		}

		hand = hand.prev

		sieve.hand = hand

		nodeToBeDeleted := hand.next
		sieve.FifoQueue.deleteNode(nodeToBeDeleted)
		delete(sieve.Nodes, nodeToBeDeleted.key)
	}

	head := sieve.FifoQueue.getHead()

	currNode := sieve.FifoQueue.insertNode(data, head, head.next)
	currNode.key = key
	sieve.Nodes[key] = currNode
	currNode.visited = false

}
