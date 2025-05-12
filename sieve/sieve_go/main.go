package sievego

type Node struct {
	end_identifier int
	next           *Node
	prev           *Node
	value          any
	visited        bool
}

func NewNode(value any) *Node {

	return &Node{
		value: value,
	}

}

type Sieve[T comparable] struct {
	hand *Node
	Capacity  int
	Nodes     map[T]*Node
	FifoQueue *FIFOQueue
}

func NewSieve[T comparable](cap int) *Sieve[T]{
	return &Sieve[T]{
		Capacity: cap,
		Nodes: make(map[T]*Node),
		FifoQueue: NewFifoQueue(),
		hand: nil,
	}
}

type FIFOQueue struct {
	head *Node
	tail *Node
}

func NewFifoQueue() *FIFOQueue {

	head := &Node{
		visited: false,
		end_identifier: 1,
	}

	tail := &Node{
		visited: false,
		end_identifier: -1,
	}
	head.next = tail
	tail.prev = head

	return &FIFOQueue{
		head: head,
		tail: tail,
	}

}

func (sieve *Sieve[T]) IsEmpty() bool {
	return len(sieve.Nodes) == 0
}

func (sieve *Sieve[T]) GetHand() *Node{
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

func (fifoQueue *FIFOQueue) GetHead() *Node {
	head:=fifoQueue.head
	if (head.end_identifier!=1){
		panic("head is not initialized properly")
	}

	return head
}

func (fifoQueue *FIFOQueue) GetTail() *Node {
	tail:= fifoQueue.tail
	if (tail.end_identifier!=-1){
		panic("tail is not initialized properly")
	}
	return tail
}


func (fifoQueue *FIFOQueue) DeleteNode(node *Node) {
	if node.end_identifier==-1 || node.end_identifier==1{
		return 
	}

	prev:=node.prev
	next:=node.next

	prev.next=next
	next.prev=prev
}


func (fifoQueue *FIFOQueue) InsertNode(value any, prev, next *Node) *Node {
	curr := NewNode(value)

	curr.next = next
	next.prev = curr

	prev.next = curr
	curr.prev = prev

	return curr
}
func (sieve *Sieve[T]) Insert(key T, data any){
	if len(sieve.Nodes) == sieve.Capacity {
		hand:=sieve.GetHand()
		if (hand==nil || hand.end_identifier==1){
			hand=sieve.FifoQueue.GetTail()
		}

		for (hand.visited && hand.end_identifier!=1){
			hand.visited=false
			hand=hand.prev
		}

		if(hand.end_identifier==1){
			hand=sieve.FifoQueue.GetTail()
		}

		hand=hand.prev
		sieve.FifoQueue.DeleteNode(hand.next)
	}

	head := sieve.FifoQueue.GetHead()
	
	currNode:=sieve.FifoQueue.InsertNode(data,head,head.next)

	currNode.visited=false

}
