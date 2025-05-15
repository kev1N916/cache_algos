package lrukgo

import (
	"sync"
	"time"
)

type LRU_K[T comparable] struct {
	K  int
	Mu sync.Mutex

	// These two data structures are maintained for all pages with a
	// Backward K-distance that is smaller than the Retained
	// Information Period. An asynchronous demon process should
	// purge history control blocks that are no longer justified
	// under the retained information criterion.
	HIST *History[T]
	LAST *Last[T]

	CRP int64
	RIP int64

	Buffer map[T][]byte

	Cap             int
	CleanupInterval time.Duration
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

func (Last *Last[T]) get(key T) int64 {
	t, present := Last.last[key]
	if !present {
		panic("key not present")
	}
	return t
}

func (Last *Last[T]) set(key T, time int64) {
	Last.last[key] = time
}

func (Last *Last[T]) delete(key T) {
	delete(Last.last, key)
}

func (Hist *History[T]) delete(key T) {
	delete(Hist.hist, key)
}

func (Hist *History[T]) get(key T, index int) int64 {
	reference_times, present := Hist.hist[key]
	if !present {
		panic("key not present")
	}

	return reference_times[index]

}

func (Hist *History[T]) init(key T, k int) {
	Hist.hist[key] = make([]int64, k)
}

func (Hist *History[T]) exists(key T) bool {
	_, present := Hist.hist[key]
	return present
}

func (Hist *History[T]) set(key T, index int, time int64) {
	_, present := Hist.hist[key]
	if !present {
		panic("key not present")
	}

	Hist.hist[key][index] = time
}

// Each time a page p is referenced, it is
// made buffer resident (it might already be buffer resident),
// and we have a history of at least one reference. If the prior
// access to page p was so long ago that we have no record of
// it, then after the Conelated Referenm Period we say that our
// estimate of bt(p,2) is infmity, and make the containing
// buffer slot available on demand. However, although we
// may drop p from memory, we need to keep history information about
// the page around for awhile; otherwise we might
// reference the page p again relatively quickly and once again
// have no record of prior reference, drop it again, reference it
// again, etc. Though the page is frequently referenced, we
// would have no history about it to recognize this fact. For
// this reason, we assume that the system will maintain history information
// about any page for some period after its
// most recent access. We refer to this period as the Retained
// Information Period.

// If a disk page p that has never been referenced before suddenly
// becomes popular enough to be kept in buffer, we
// should recognize this fact as long as two references to the
// page are no more thau the Retained Information Period
// apart. Though we drop the page after the first reference, we keep information
// around in memory to recognize when a
// second reference gives a value of bt(p,2) that passes our
// LRU-2 criterion for retention in buffer. The page history information
// kept in a memory resident data structure is designated by HIST(p),
// and contains the last two reference
// string subscripts i and j, where ri = rj = p, or just the last
// reference if only one is known.
func (lru *LRU_K[T]) FindVictim(t int64) T {
	min := t
	var victim T
	for page := range lru.Buffer {
		time_of_last_reference := lru.LAST.get(page)
		if t-time_of_last_reference > lru.CRP && lru.HIST.get(page, lru.K-1) < min {
			victim = page
			min = lru.HIST.get(page, lru.K-1)
		}
	}

	return victim
}

func NewLRU[T comparable](k int, cap int, crp int64) *LRU_K[T] {
	last := NewLast[T]()
	history := NewHistory[T](k)

	lru_k := &LRU_K[T]{
		Mu:              sync.Mutex{},
		K:               k,
		CRP:             crp,
		Cap:             cap,
		LAST:            last,
		HIST:            history,
		CleanupInterval: 2 * time.Minute,
		Buffer: make(map[T][]byte),
	}
	// lru_k.Buffer=make(map[T][]byte,cap)
	return lru_k
}

func (lru *LRU_K[T]) Get(key T) ([]byte, bool) {
	lru.Mu.Lock()
	defer lru.Mu.Unlock()

	data, present := lru.Buffer[key]
	return data, present
}

func (lru *LRU_K[T]) Cleanup(key T) {
	lru.Mu.Lock()
	defer lru.Mu.Unlock()

	delete(lru.Buffer, key)
	lru.HIST.delete(key)
	lru.LAST.delete(key)
}

// These two data structures are maintained for all pages with a
// Backward K-distance that is smaller than the Retained
// Information Period. An asynchronous demon process should
// purge history control blocks that are no longer justified under
// the retained information criterion.
func (lru *LRU_K[T]) StartCleanup() {
	cleanupInterval := lru.CleanupInterval
	for {
		time.Sleep(cleanupInterval)

		for page := range lru.Buffer {
			lru.Mu.Lock()
			backward_K_Distance := lru.HIST.get(page, lru.K-1)
			lru.Mu.Unlock()
			if backward_K_Distance > lru.RIP {
				go lru.Cleanup(page)
			}
		}
	}
}
func (lru *LRU_K[T]) Set(key T, data []byte) (success bool) {
	lru.Mu.Lock()
	defer lru.Mu.Unlock()

	t := time.Now().Unix()
	_, present := lru.Buffer[key]
	if present {
		time_of_last_reference := lru.LAST.get(key)

		// The system should not drop a page immediately after
		// its first reference, but should keep the page around for a
		// short period until the likelihood of a dependent follow-up
		// reference is minimal; then the page can be dropped.
		// At the same time, interarrival time should be calculated based
		// on non-correlated access pairs, where each successive access by
		// the same process within a time-out period is assumed to be correlated
		// the relationship is transitive. We refer to this approach, which associates
		// correlated references, as the Time-Out Correlation method;
		// and we refer to the time-out period as the Correlated Reference Period.
		//
		// If a reference to a page p is made several
		// times during a Correlated Reference Period, we do not
		//  want to penalize or credit the page for that.
		if t-time_of_last_reference > lru.CRP {
			correl_period_of_refd_page := lru.LAST.get(key) - lru.HIST.get(key, 0)

			for i := 1; i < lru.K; i++ {
				prev_reference_time := lru.HIST.get(key, i-1)
				lru.HIST.set(key, i, prev_reference_time+correl_period_of_refd_page)
			}

			lru.HIST.set(key, 0, t)
			lru.LAST.set(key, t)
		} else {
			lru.LAST.set(key, t)
		}

		lru.Buffer[key] = data
	} else {
		if len(lru.Buffer) < lru.Cap {
			lru.Buffer[key] = data

			lru.LAST.set(key, t)
			lru.HIST.init(key, lru.K)
			lru.HIST.set(key, 0, t)

		} else if(len(lru.Buffer)>=lru.Cap) {
			victim := lru.FindVictim(t)
			
			delete(lru.Buffer, victim)
			lru.LAST.delete(victim)

			lru.Buffer[key] = data
			if !lru.HIST.exists(key) {
				lru.HIST.init(key, lru.K)
			} else {
				for i := 1; i < lru.K; i++ {
					prev_reference_time := lru.HIST.get(key, i-1)
					lru.HIST.set(key, i, prev_reference_time)
				}
			}

			lru.HIST.set(key, 0, t)
			lru.LAST.set(key, t)
		}

	}
	return true

}
