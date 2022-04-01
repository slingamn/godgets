// Copyright (c) 2022 Shivaram Lingamneni
// released under the 0BSD license

package godgets

/*
This is a type-safe generic slab-allocated LRU cache:

1. Generic as in Go 1.18 type parameters
2. Slab-allocated as in the linked list is implemented with integer indices
into a contiguous slice of nodes, as opposed to pointers (like container/list
or zyedidia/generic). This theoretically increases GC performance since the
GC has fewer total allocations to track. See here:
https://github.com/kentik/patricia
*/

type LRU[K comparable, V any] struct {
	maxSize int
	// map keys to their index in the slab array:
	items map[K]int
	// slab array of key-value pairs, doubly-linked in access order:
	slab []Node[K, V]
	// front and back of the list (as indices in the slab array, -1 for nonexistent)
	front int
	back  int
	// indices that were Remove()'d and can be used for new allocations:
	freeList []int

	onEvict LRUCallback[K, V]
}

type Node[K comparable, V any] struct {
	Key   K
	Value V
	// previous and next "pointers" (indices in the slab array)
	// -1 for no such element; the zero value (0, 0) is considered
	// equivalent to (-1, -1)
	prev int
	next int
}

type LRUCallback[K comparable, V any] func(key K, value V)

func (lru *LRU[K, V]) Initialize(initialSize, maxSize int, onEvict LRUCallback[K, V]) {
	lru.maxSize = maxSize
	lru.onEvict = onEvict
	lru.items = make(map[K]int, initialSize)
	lru.slab = make([]Node[K, V], 0, initialSize)
	lru.front = -1
	lru.back = -1

	lru.freeList = nil
}

func NewLRU[K comparable, V any](initialSize, maxSize int, onEvict LRUCallback[K, V]) *LRU[K, V] {
	result := new(LRU[K, V])
	result.Initialize(initialSize, maxSize, onEvict)
	return result
}

func (c *LRU[K, V]) Purge() {
	idx := c.back
	for idx != -1 {
		if c.onEvict != nil {
			c.onEvict(c.slab[idx].Key, c.slab[idx].Value)
		}
		nextIdx := c.slab[idx].next
		delete(c.items, c.slab[idx].Key)
		idx = nextIdx
	}
	c.slab = c.slab[:0]
	c.front = -1
	c.back = -1
	c.freeList = c.freeList[:0]
}

func (c *LRU[K, V]) Add(key K, value V) (evicted bool) {
	if idx, found := c.items[key]; found {
		// found existing item
		c.slab[idx].Value = value
		c.moveToFront(idx)
		return false
	}

	var idx int
	if len(c.freeList) != 0 {
		// pop from free list
		idx = c.freeList[len(c.freeList)-1]
		c.freeList = c.freeList[:len(c.freeList)-1]
	} else if len(c.slab) < cap(c.slab) || cap(c.slab) < c.maxSize {
		// allocate a new entry in the slab
		c.growSlab()
		idx = len(c.slab)
		c.slab = c.slab[:idx+1]
	} else {
		// eviction
		idx = c.back
		delete(c.items, c.slab[idx].Key)
		if c.onEvict != nil {
			c.onEvict(c.slab[idx].Key, c.slab[idx].Value)
		}
		evicted = true
	}

	c.slab[idx].Key = key
	c.slab[idx].Value = value
	c.items[key] = idx
	c.moveToFront(idx)
	return
}

func (c *LRU[K, V]) Get(key K) (value V, ok bool) {
	if idx, ok := c.items[key]; ok {
		c.moveToFront(idx)
		return c.slab[idx].Value, true
	}
	return
}

func (c *LRU[K, V]) Contains(key K) (ok bool) {
	_, ok = c.items[key]
	return ok
}

func (c *LRU[K, V]) Peek(key K) (value V, ok bool) {
	if idx, ok := c.items[key]; ok {
		return c.slab[idx].Value, true
	}
	return
}

func (c *LRU[K, V]) Remove(key K) (present bool) {
	if idx, ok := c.items[key]; ok {
		delete(c.items, key)
		prev := c.slab[idx].prev
		next := c.slab[idx].next
		if c.front == idx {
			c.front = prev
		}
		if c.back == idx {
			c.back = next
		}
		if prev != -1 {
			c.slab[prev].next = next
		}
		if next != -1 {
			c.slab[next].prev = prev
		}
		if c.onEvict != nil {
			c.onEvict(key, c.slab[idx].Value)
		}
		c.slab[idx] = Node[K, V]{}
		c.freeList = append(c.freeList, idx)
		return true
	}
	return false
}

func (c *LRU[K, V]) Iterate(callback LRUCallback[K, V]) {
	for idx := c.back; idx != -1; idx = c.slab[idx].next {
		callback(c.slab[idx].Key, c.slab[idx].Value)
	}
}

func (c *LRU[K, V]) Len() int {
	return len(c.items)
}

func (c *LRU[K, V]) moveToFront(idx int) {
	prev, next := c.slab[idx].prev, c.slab[idx].next
	if prev == 0 && next == 0 {
		// freshly allocated or from the free list, invalid:
		prev, next = -1, -1
	}
	if prev != -1 {
		c.slab[prev].next = next
	}
	if next != -1 {
		c.slab[next].prev = prev
	}
	c.slab[idx].next = -1
	if c.front != idx {
		c.slab[idx].prev = c.front
		if c.front != -1 {
			c.slab[c.front].next = idx
		}
		c.front = idx
	}
	if c.back == -1 {
		c.back = idx
	} else if c.back == idx && next != -1 {
		c.back = next
	}
}

func (c *LRU[K, V]) growSlab() {
	if len(c.slab) < cap(c.slab) {
		return
	}
	size := cap(c.slab) * 2
	if size == 0 {
		size = 1
	} else if size > c.maxSize {
		size = c.maxSize
	}
	slab := make([]Node[K, V], len(c.slab), size)
	copy(slab, c.slab)
	c.slab = slab
}
