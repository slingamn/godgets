// Copyright (c) 2019 Shivaram Lingamneni
// released under the MIT license

package godgets

import (
	"testing"
	"time"
)

func TestTryAcquire(t *testing.T) {
	count := 3
	sem := NewSemaphore(count)

	for i := 0; i < count; i++ {
		assertEqual(sem.TryAcquire(), true)
	}
	// used up the capacity
	assertEqual(sem.TryAcquire(), false)
	sem.Release()
	// got one slot back
	assertEqual(sem.TryAcquire(), true)
}

func TestAcquireWithTimeout(t *testing.T) {
	sem := NewSemaphore(1)

	assertEqual(sem.TryAcquire(), true)

	// cannot acquire the held semaphore
	assertEqual(sem.AcquireWithTimeout(100*time.Millisecond), false)

	sem.Release()
	// can acquire the released semaphore
	assertEqual(sem.AcquireWithTimeout(100*time.Millisecond), true)
	sem.Release()

	// XXX this test could fail if the machine is extremely overloaded
	sem.Acquire()
	go func() {
		time.Sleep(100 * time.Millisecond)
		sem.Release()
	}()
	// we should acquire successfully after approximately 100 msec
	assertEqual(sem.AcquireWithTimeout(1*time.Second), true)
}
