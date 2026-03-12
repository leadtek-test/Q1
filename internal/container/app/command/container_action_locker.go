package command

import (
	"context"
	stderrors "errors"
	"fmt"
	"sync"
	"time"
)

type ContainerActionLocker interface {
	Lock(ctx context.Context, userID, containerID uint) (unlock func(), waited bool, err error)
}

var ErrContainerActionWaitTimeout = stderrors.New("container action wait timeout")

var defaultContainerActionLocker ContainerActionLocker = NewInMemoryContainerActionLocker()

type inMemoryContainerActionLocker struct {
	mu          sync.Mutex
	locks       map[string]*containerActionLockState
	waitTimeout time.Duration
	holdTimeout time.Duration
}

type containerActionLockState struct {
	waiters     []*containerActionWaiter
	holder      *containerActionWaiter
	holderTimer *time.Timer
}

type containerActionWaiter struct {
	ready   chan struct{}
	granted bool
}

func NewInMemoryContainerActionLocker() ContainerActionLocker {
	return newInMemoryContainerActionLocker(3*time.Second, 10*time.Second)
}

func newInMemoryContainerActionLocker(waitTimeout, holdTimeout time.Duration) ContainerActionLocker {
	if waitTimeout <= 0 {
		panic("container action lock wait timeout must be positive")
	}
	if holdTimeout <= 0 {
		panic("container action lock hold timeout must be positive")
	}
	return &inMemoryContainerActionLocker{
		locks:       make(map[string]*containerActionLockState),
		waitTimeout: waitTimeout,
		holdTimeout: holdTimeout,
	}
}

func (l *inMemoryContainerActionLocker) Lock(ctx context.Context, userID, containerID uint) (func(), bool, error) {
	key := fmt.Sprintf("%d:%d", userID, containerID)
	waiter := &containerActionWaiter{ready: make(chan struct{})}

	l.mu.Lock()
	state := l.locks[key]
	if state == nil {
		state = &containerActionLockState{}
		l.locks[key] = state
	}
	waited := len(state.waiters) > 0
	state.waiters = append(state.waiters, waiter)
	if !waited {
		l.grantHeadLocked(key, state)
	}
	l.mu.Unlock()

	if waited {
		if err := l.waitForGrant(ctx, key, waiter); err != nil {
			return nil, waited, err
		}
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			l.release(key, waiter)
		})
	}, waited, nil
}

func (l *inMemoryContainerActionLocker) waitForGrant(ctx context.Context, key string, waiter *containerActionWaiter) error {
	timer := time.NewTimer(l.waitTimeout)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	select {
	case <-waiter.ready:
		return nil
	case <-ctx.Done():
		if l.cancelWaitingWaiter(key, waiter) {
			return ctx.Err()
		}
		<-waiter.ready
		return nil
	case <-timer.C:
		if l.cancelWaitingWaiter(key, waiter) {
			return ErrContainerActionWaitTimeout
		}
		<-waiter.ready
		return nil
	}
}

func (l *inMemoryContainerActionLocker) cancelWaitingWaiter(key string, waiter *containerActionWaiter) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	state := l.locks[key]
	if state == nil {
		return false
	}

	idx := indexWaiter(state.waiters, waiter)
	if idx == -1 {
		return false
	}

	target := state.waiters[idx]
	if target.granted {
		return false
	}

	state.waiters = append(state.waiters[:idx], state.waiters[idx+1:]...)
	if len(state.waiters) == 0 {
		if state.holderTimer != nil {
			state.holderTimer.Stop()
			state.holderTimer = nil
		}
		state.holder = nil
		delete(l.locks, key)
	}
	return true
}

func (l *inMemoryContainerActionLocker) grantHeadLocked(key string, state *containerActionLockState) {
	if len(state.waiters) == 0 {
		state.holder = nil
		return
	}

	head := state.waiters[0]
	state.holder = head
	if !head.granted {
		head.granted = true
		close(head.ready)
	}
	if state.holderTimer != nil {
		state.holderTimer.Stop()
	}
	state.holderTimer = time.AfterFunc(l.holdTimeout, func() {
		l.forceRelease(key, head)
	})
}

func (l *inMemoryContainerActionLocker) release(key string, waiter *containerActionWaiter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	state := l.locks[key]
	if state == nil {
		return
	}
	if state.holder != waiter {
		return
	}

	l.releaseHolderLocked(key, state, waiter)
}

func (l *inMemoryContainerActionLocker) forceRelease(key string, waiter *containerActionWaiter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	state := l.locks[key]
	if state == nil {
		return
	}
	if state.holder != waiter {
		return
	}

	l.releaseHolderLocked(key, state, waiter)
}

func (l *inMemoryContainerActionLocker) releaseHolderLocked(key string, state *containerActionLockState, waiter *containerActionWaiter) {
	if state.holderTimer != nil {
		state.holderTimer.Stop()
		state.holderTimer = nil
	}
	state.holder = nil

	idx := indexWaiter(state.waiters, waiter)
	if idx >= 0 {
		state.waiters = append(state.waiters[:idx], state.waiters[idx+1:]...)
	}

	if len(state.waiters) == 0 {
		delete(l.locks, key)
		return
	}

	l.grantHeadLocked(key, state)
}

func indexWaiter(waiters []*containerActionWaiter, target *containerActionWaiter) int {
	for idx, waiter := range waiters {
		if waiter == target {
			return idx
		}
	}
	return -1
}
