package fakes

import "sync"

// Counterfeiter does not lock "*Returns(...)" functions
// leading to a race, so we had to hand-roll this fake
type FakeHealthChecker struct {
	sync.RWMutex
	status bool
}

func (checker *FakeHealthChecker) Check() bool {
	checker.RLock()
	defer checker.RUnlock()

	return checker.status
}

func (checker *FakeHealthChecker) CheckReturns(status bool) {
	checker.Lock()
	defer checker.Unlock()

	checker.status = status
}

func NewFakeHealthChecker() *FakeHealthChecker {
	return &FakeHealthChecker{
		status: false,
	}
}
