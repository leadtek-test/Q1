package command

import (
	"context"
	stderrors "errors"
	"testing"
	"time"
)

func TestInMemoryContainerActionLockerWaitTimeout(t *testing.T) {
	locker := newInMemoryContainerActionLocker(80*time.Millisecond, time.Second)

	unlock, waited, err := locker.Lock(context.Background(), 1, 100)
	if err != nil || waited {
		t.Fatalf("first lock should acquire immediately, waited=%v err=%v", waited, err)
	}
	defer unlock()

	start := time.Now()
	_, waited, err = locker.Lock(context.Background(), 1, 100)
	if !waited {
		t.Fatalf("second lock should be queued")
	}
	if !stderrors.Is(err, ErrContainerActionWaitTimeout) {
		t.Fatalf("expected wait timeout error, got=%v", err)
	}
	if time.Since(start) < 70*time.Millisecond {
		t.Fatalf("lock timeout happened too early: %s", time.Since(start))
	}
}

func TestInMemoryContainerActionLockerAutoReleaseAfterHoldTimeout(t *testing.T) {
	locker := newInMemoryContainerActionLocker(2*time.Second, 120*time.Millisecond)

	unlock1, waited, err := locker.Lock(context.Background(), 9, 9)
	if err != nil || waited {
		t.Fatalf("first lock should acquire immediately, waited=%v err=%v", waited, err)
	}

	type result struct {
		waited bool
		err    error
		unlock func()
	}
	done := make(chan result, 1)
	go func() {
		unlock2, waited2, err2 := locker.Lock(context.Background(), 9, 9)
		done <- result{waited: waited2, err: err2, unlock: unlock2}
	}()

	r := <-done
	if r.err != nil {
		t.Fatalf("second lock should acquire after auto-release, err=%v", r.err)
	}
	if !r.waited {
		t.Fatalf("second lock should have waited")
	}
	if r.unlock == nil {
		t.Fatal("second lock should return unlock function")
	}
	r.unlock()

	// should be idempotent/no-op even after auto-release.
	unlock1()
}
