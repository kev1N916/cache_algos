package lrukgo

import (
	"sync"
	"time"
)

type LRU_K[T comparable] struct {
	K  int
	Mu sync.Mutex

	HIST *History[T]

	CRP  int64
	RIP int64 
	LAST *Last[T]

	Buffer map[T][]byte

	Cap int
}

type Last[T comparable] struct {
	last map[T]int64
}

type History[T comparable] struct {
	hist map[T][]int64
}

func NewLast[T comparable]() *Last[T] {
	last := &Last[T]{
		last: make(map[T]int64),
	}
	return last
}

func NewHistory[T comparable](k int) *History[T] {
	history := &History[T]{
		hist: make(map[T][]int64),
	}
	return history
}

func (Last *Last[T]) Get(key T) int64 {
	t, present := Last.last[key]
	if !present {
		panic("key not present")
	}
	return t
}

func (Last *Last[T]) Set(key T, time int64) {
	Last.last[key] = time
}

func (Hist *History[T]) Get(key T, index int) int64 {
	reference_times, present := Hist.hist[key]
	if !present {
		panic("key not present")
	}

	return reference_times[index]

}

func (Hist *History[T]) Set(key T, index int, time int64) {
	_, present := Hist.hist[key]
	if !present {
		panic("key not present")
	}

	Hist.hist[key][index] = time
}

func (lru *LRU_K[T]) FindVictim(t int64) T {
	min := t
	var victim T
	for page := range lru.Buffer {
		time_of_last_reference := lru.LAST.Get(page)
		if t-time_of_last_reference > lru.CRP && lru.HIST.Get(page,lru.K-1) < min {
			victim = page
			min = lru.HIST.Get(page,lru.K-1)
		}
	}

	return victim
}

func NewLRU[T comparable](k int, cap int, crp int64) *LRU_K[T] {
	last := NewLast[T]()
	history := NewHistory[T](k)

	lru_k := &LRU_K[T]{
		Mu:   sync.Mutex{},
		K:    k,
		CRP:  crp,
		Cap:  cap,
		LAST: last,
		HIST: history,
	}
	return lru_k
}

// Each time a page p is referenced, it is
// made buffer resident (it might already be buffer resident),
// and we have a history of at least one reference. If the prior
// access to page p was so long ago that we have no record of
// it, then after the Conelated Referenm Period we say that our
// estimate of bt(p,2) is infhit y, and make the containing
// buffer slot available on demand. However, although we
// may drop p from memory, we need to keep history information about 
// the page around for awhile; otherwise we might
// reference the page p again relatively qnicld y and once again
// have no record of prior reference, drop it again, reference it
// again, etc. Though the page is frequently referenced, we
// would have no history about it to recognize this fact. For
// this reason, we assume that the system will maintain history information
// about any page for some period after its
// most recent access. We refer to this period as the Retained
// Information Period.

// If a disk page p that has never been referenced before sud
// denly becomes popular enough to be kept in buffer, we
// should recognize this fact as long as two references to the
// page are no more thau the Retained Information Period
// apart. Though we drop the page after the first reference, we keep information
// around in memory to recognize when a
// second reference gives a value of bt(p,2) that passes our
// LRU-2 criterion for retention in buffer. The page history information
// kept in a memory resident data structure is designated by HIST(p), 
// and contains the last two reference
// string subscripts i and j, where ri = rj = p, or just the last
// reference if onl y one is known. 
func (lru *LRU_K[T]) Set(key T, data []byte) (success bool) {
	lru.Mu.Lock()
	defer lru.Mu.Unlock()

	t := time.Now().Unix()
	_, present := lru.Buffer[key]
	if present {
		time_of_last_reference := lru.LAST.Get(key)

		if t-time_of_last_reference > lru.CRP {
			correl_period_of_refd_page := lru.LAST.Get(key) - lru.HIST.Get(key, 0)

			for i := 1; i < lru.K; i++ {
				// lru.HIST.hist[key][i] = lru.HIST.hist[key][i-1] + correl_period_of_refd_page
				prev_reference_time := lru.HIST.Get(key, i-1)
				lru.HIST.Set(key, i, prev_reference_time+correl_period_of_refd_page)
			}

			// lru.HIST.hist[key][0] = t
			// lru.LAST.last[key] = t

			lru.HIST.Set(key, 0, t)
			lru.LAST.Set(key, t)

		} else {
			lru.LAST.Set(key, t)
		}

		lru.Buffer[key] = data
	} else {
		if len(lru.Buffer) < lru.Cap {
			lru.Buffer[key] = data
			// lru.LAST.last[key] = t
			// lru.HIST.hist[key] = make([]int64, lru.K)
			// lru.HIST.hist[key][0] = t
		} else {
			victim:=lru.FindVictim(t)
			delete(lru.Buffer, victim)

		}

	}
	return true

}
