package failover_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
)

func TestFailover(t *testing.T) {
	helpers.PrepareAndRunTests("Failover", t, true)
}
