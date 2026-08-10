package safego

import (
	"testing"
	"time"
)

func TestGoContainsPanic(t *testing.T) {
	done := make(chan struct{})
	Go(func() {
		defer close(done)
		panic("boom")
	})
	select {
	case <-done:
		// Panic was contained and the test process survived.
	case <-time.After(2 * time.Second):
		t.Fatal("safego goroutine did not finish")
	}
}

func TestGoRunsNormally(t *testing.T) {
	done := make(chan struct{})
	Go(func() {
		defer close(done)
	})
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("safego goroutine did not run")
	}
}
