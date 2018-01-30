package healthchecker

import ()

type CompositeChecker struct {
	healthCheckers []HealthChecker
}

func NewCompositeChecker(checkers []HealthChecker) *CompositeChecker {
	return &CompositeChecker{
		healthCheckers: checkers,
	}
}

func (compositeChecker *CompositeChecker) Check() bool {
	status := true
	for _, checker := range compositeChecker.healthCheckers {
		status = status && checker.Check()
	}
	return status
}
