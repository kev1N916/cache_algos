package lfuo1

type FreqNode[T comparable] struct {
	value int
	items map[T]*LFU_Item[T]
	prev  *FreqNode[T]
	next  *FreqNode[T]
}

func NewFreqNode[T comparable]() *FreqNode[T] {

	return &FreqNode[T]{
		value: 0,
		items: make(map[T]*LFU_Item[T]),
		prev:  nil,
		next:  nil,
	}
}

func GetNewNode[T comparable](value int, prev, next *FreqNode[T]) *FreqNode[T] {
	new_node := NewFreqNode[T]()
	new_node.value = value
	new_node.prev = prev
	new_node.next = next
	prev.next = new_node
	next.prev = new_node
	return new_node
}

func DeleteNode[T comparable](node *FreqNode[T]) {
	next := node.next
	prev := node.prev

	prev.next = next
	next.prev = prev
}

type LFU_Item[T comparable] struct {
	data   any
	parent *FreqNode[T]
}

func NewLfuItem[T comparable](data any, parent *FreqNode[T]) *LFU_Item[T] {

	return &LFU_Item[T]{
		data:   data,
		parent: parent,
	}
}

type LFU_Cache[T comparable] struct {
	size int 
	bykey     map[T]*LFU_Item[T]
	freq_Head *FreqNode[T]
}

func NewLfuCache[T comparable]() *LFU_Cache[T] {

	return &LFU_Cache[T]{
		bykey: make(map[T]*LFU_Item[T]),
		freq_Head: &FreqNode[T]{
			value: 0,
			prev:  nil,
			next:  nil,
		},
	}
}

func (lfuCache *LFU_Cache[T]) Insert(key T, value any) {

	_, present := lfuCache.bykey[key]
	if present {
		panic("Key already exists")
	}

	if len(lfuCache.bykey)==lfuCache.size{
		lfuCache.Evict()
	}

	freq := lfuCache.freq_Head.next
	if freq == nil {
		freq = GetNewNode(1, lfuCache.freq_Head, freq)
	} else if freq.value != 1 {
		freq = GetNewNode(1, lfuCache.freq_Head, freq)
	}

	lfuItem := NewLfuItem(value, freq)
	lfuCache.bykey[key] = lfuItem
	freq.items[key] = lfuItem
}

func (lfuCache *LFU_Cache[T]) Access(key T) (value any) {

	tmp := lfuCache.bykey[key]
	if tmp == nil {
		panic("No such key")
	}

	freq := tmp.parent
	next_freq := freq.next
	if next_freq == nil || next_freq.value != freq.value+1 {
		next_freq = GetNewNode(freq.value+1, freq, next_freq)
	}

	next_freq.items[key] = tmp
	tmp.parent = next_freq

	delete(freq.items, key)
	if len(freq.items) == 0 {
		DeleteNode(freq)
	}

	return tmp.data
}

func (lfuCache *LFU_Cache[T]) Evict() (T, any) {

	var zeroValue T

	if len(lfuCache.bykey) == 0 {
		panic("the set is empty")
	}

	if lfuCache.freq_Head.next == nil {
		panic("the set is empty")
	}

	for item, present := range lfuCache.freq_Head.next.items {
		delete(lfuCache.freq_Head.next.items, item)
		delete(lfuCache.bykey, item)

		if len(lfuCache.freq_Head.next.items) == 0 {
			DeleteNode(lfuCache.freq_Head.next)
		}
		return item, present.data
	}

	return zeroValue, nil

}
