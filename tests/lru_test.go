package godgets

/*
Unlike the production code in this repository, this test code is copyright Hashicorp
and associated contributors and released under the Mozilla Public License 2.0.
*/

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func BenchmarkLRU_Rand(b *testing.B) {
	var l LRU[int64, int64]
	l.Initialize(0, 8192, nil)

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		trace[i] = rand.Int63() % 32768
	}

	b.ResetTimer()

	var hit, miss int
	for i := 0; i < 2*b.N; i++ {
		if i%2 == 0 {
			l.Add(trace[i], trace[i])
		} else {
			_, ok := l.Get(trace[i])
			if ok {
				hit++
			} else {
				miss++
			}
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

func BenchmarkLRU_Freq(b *testing.B) {
	var l LRU[int64, int64]
	l.Initialize(0, 8192, nil)

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		if i%2 == 0 {
			trace[i] = rand.Int63() % 16384
		} else {
			trace[i] = rand.Int63() % 32768
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Add(trace[i], trace[i])
	}
	var hit, miss int
	for i := 0; i < b.N; i++ {
		_, ok := l.Get(trace[i])
		if ok {
			hit++
		} else {
			miss++
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(miss))
}

// TODO will this get compiled into clients?
func (c *LRU[K, V]) integrityCheck() {
	count := 0
	for idx := c.back; idx != -1; idx = c.slab[idx].next {
		count++
		if count > len(c.items) {
			panic(fmt.Sprintf("excess or loop detected: map has %d, list has at least %d", len(c.items), count))
		}
		if c.items[c.slab[idx].Key] != idx {
			panic(fmt.Sprintf("inconsistent mapping: %v %d %d", c.slab[idx].Key, c.items[c.slab[idx].Key], idx))
		}
	}
	if count != len(c.items) {
		panic(fmt.Sprintf("undercount detected: map has %d, list has %d", len(c.items), count))
	}
	assertEqual(count == 0, c.front == -1)
	assertEqual(count == 0, c.back == -1)
	//fmt.Printf("integrity check passed: %d %d\n", c.back, len(c.items))
}

func (c *LRU[K, V]) keys() (result []K) {
	result = make([]K, 0, c.Len())
	c.Iterate(func(k K, v V) {
		result = append(result, k)
	})
	return
}

func TestLRU(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k int, v int) {
		if k != v {
			t.Fatalf("Evict values not equal (%v!=%v)", k, v)
		}
		evictCounter++
	}
	var l LRU[int, int]
	l.Initialize(0, 128, onEvicted)

	for i := 0; i < 256; i++ {
		l.Add(i, i)
		l.integrityCheck()
	}
	if l.Len() != 128 {
		t.Fatalf("bad len: %v", l.Len())
	}

	if evictCounter != 128 {
		t.Fatalf("bad evict count: %v", evictCounter)
	}

	for i, k := range l.keys() {
		if v, ok := l.Get(k); !ok || v != k || v != i+128 {
			t.Fatalf("bad key: %v", k)
		}
		l.integrityCheck()
	}
	for i := 0; i < 128; i++ {
		_, ok := l.Get(i)
		l.integrityCheck()
		if ok {
			t.Fatalf("%d should be evicted", i)
		}
	}
	for i := 128; i < 256; i++ {
		_, ok := l.Get(i)
		l.integrityCheck()
		if !ok {
			t.Fatalf("should not be evicted")
		}
	}
	for i := 128; i < 192; i++ {
		l.Remove(i)
		l.integrityCheck()
		_, ok := l.Get(i)
		if ok {
			t.Fatalf("should be deleted")
		}
	}

	val, ok := l.Get(192) // expect 192 to be last key in l.Keys()
	l.integrityCheck()
	assertEqual(val, 192)
	assertEqual(ok, true)

	for i, k := range l.keys() {
		if (i < 63 && k != i+193) || (i == 63 && k != 192) {
			t.Fatalf("out of order key: %v", k)
		}
	}

	l.Purge()
	l.integrityCheck()
	if l.Len() != 0 {
		t.Fatalf("bad len: %v", l.Len())
	}
	if _, ok := l.Get(200); ok {
		t.Fatalf("should contain nothing")
	}
	l.integrityCheck()
}

// test that Add returns true/false if an eviction occurred
func TestLRUAdd(t *testing.T) {
	evictCounter := 0
	onEvicted := func(k, v int) {
		evictCounter++
	}

	var l LRU[int, int]
	l.Initialize(1, 1, onEvicted)
	l.integrityCheck()

	if l.Add(1, 1) == true || evictCounter != 0 {
		t.Errorf("should not have an eviction")
	}
	l.integrityCheck()
	if l.Add(2, 2) == false || evictCounter != 1 {
		t.Errorf("should have an eviction")
	}
	l.integrityCheck()
}

// test that Contains doesn't update recent-ness
func TestLRUContains(t *testing.T) {
	var l LRU[int, int]
	l.Initialize(2, 2, nil)

	l.Add(1, 1)
	l.integrityCheck()
	l.Add(2, 2)
	l.integrityCheck()
	if !l.Contains(1) {
		t.Errorf("1 should be contained")
	}

	l.Add(3, 3)
	l.integrityCheck()
	if l.Contains(1) {
		t.Errorf("Contains should not have updated recent-ness of 1")
	}
}

// test that Peek doesn't update recent-ness
func TestLRUPeek(t *testing.T) {
	var l LRU[int, int]
	l.Initialize(2, 2, nil)
	l.integrityCheck()

	l.Add(1, 1)
	l.integrityCheck()
	l.Add(2, 2)
	l.integrityCheck()
	if v, ok := l.Peek(1); !ok || v != 1 {
		t.Errorf("1 should be set to 1: %v, %v", v, ok)
	}

	l.Add(3, 3)
	l.integrityCheck()
	if l.Contains(1) {
		t.Errorf("should not have updated recent-ness of 1")
	}
	if !l.Contains(2) {
		t.Errorf("2 shouldn't have been affected")
	}
}

func assertEqual(found, expected interface{}) {
	if !reflect.DeepEqual(found, expected) {
		panic(fmt.Sprintf("found %#v, expected %#v", found, expected))
	}
}

func TestLRURemove(t *testing.T) {
	var l LRU[int, int]
	l.Initialize(1, 1, nil)
	l.Add(1, 1)
	l.integrityCheck()
	v, ok := l.Get(1)
	l.integrityCheck()
	assertEqual(v, 1)
	assertEqual(ok, true)
	v, ok = l.Get(1)
	l.integrityCheck()
	assertEqual(v, 1)
	assertEqual(ok, true)
	removed := l.Remove(1)
	l.integrityCheck()
	assertEqual(removed, true)
	removed = l.Remove(2)
	l.integrityCheck()
	assertEqual(removed, false)

	evicted := l.Add(2, 2)
	l.integrityCheck()
	assertEqual(evicted, false)
	v, ok = l.Get(2)
	assertEqual(v, 2)
	assertEqual(ok, true)
}

func TestLRURemove2(t *testing.T) {
	l := NewLRU[int, int](2, 2, nil)
	l.Add(1, 1)
	l.integrityCheck()
	v, ok := l.Get(1)
	l.integrityCheck()
	assertEqual(v, 1)
	assertEqual(ok, true)
	l.Add(2, 2)
	l.integrityCheck()
	assertEqual(l.Contains(2), true)
	removed := l.Remove(1)
	l.integrityCheck()
	assertEqual(removed, true)
	assertEqual(l.Contains(2), true)
	removed = l.Remove(2)
	l.integrityCheck()
	assertEqual(removed, true)
	assertEqual(l.Contains(2), false)
}

func TestOrdering(t *testing.T) {
	var l LRU[int, int]
	l.Initialize(8, 8, nil)

	check := func() {
		l.integrityCheck()
		for _, k := range l.keys() {
			v, ok := l.Peek(k)
			assertEqual(ok, true)
			assertEqual(k, v)
		}
	}

	for i := 0; i < 8; i++ {
		l.Add(i, i)
		check()
	}

	assertEqual(l.keys(), []int{0, 1, 2, 3, 4, 5, 6, 7})

	for i := 0; i < 8; i++ {
		if i%2 == 0 {
			l.Get(i)
			check()
		}
	}

	assertEqual(l.keys(), []int{1, 3, 5, 7, 0, 2, 4, 6})

	l.Add(8, 8)
	check()

	assertEqual(l.keys(), []int{3, 5, 7, 0, 2, 4, 6, 8})

	l.Remove(5)
	check()
	l.Remove(7)
	check()
	l.Remove(6)
	check()
	l.Remove(0)
	check()
	assertEqual(l.keys(), []int{3, 2, 4, 8})
	l.Remove(3)
	check()
	assertEqual(l.keys(), []int{2, 4, 8})
	l.Remove(8)
	check()
	assertEqual(l.keys(), []int{2, 4})
}
