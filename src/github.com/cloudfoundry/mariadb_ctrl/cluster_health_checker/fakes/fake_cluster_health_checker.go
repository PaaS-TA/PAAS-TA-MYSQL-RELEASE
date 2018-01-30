// This file was generated by counterfeiter
package fakes

import (
	"sync"

	"github.com/cloudfoundry/mariadb_ctrl/cluster_health_checker"
)

type FakeClusterHealthChecker struct {
	HealthyClusterStub        func() bool
	healthyClusterMutex       sync.RWMutex
	healthyClusterArgsForCall []struct{}
	healthyClusterReturns     struct {
		result1 bool
	}
}

func (fake *FakeClusterHealthChecker) HealthyCluster() bool {
	fake.healthyClusterMutex.Lock()
	fake.healthyClusterArgsForCall = append(fake.healthyClusterArgsForCall, struct{}{})
	fake.healthyClusterMutex.Unlock()
	if fake.HealthyClusterStub != nil {
		return fake.HealthyClusterStub()
	} else {
		return fake.healthyClusterReturns.result1
	}
}

func (fake *FakeClusterHealthChecker) HealthyClusterCallCount() int {
	fake.healthyClusterMutex.RLock()
	defer fake.healthyClusterMutex.RUnlock()
	return len(fake.healthyClusterArgsForCall)
}

func (fake *FakeClusterHealthChecker) HealthyClusterReturns(result1 bool) {
	fake.HealthyClusterStub = nil
	fake.healthyClusterReturns = struct {
		result1 bool
	}{result1}
}

var _ cluster_health_checker.ClusterHealthChecker = new(FakeClusterHealthChecker)
