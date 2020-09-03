package dashboard_test

import (
	"testing"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
)

func TestDashboard(t *testing.T) {
	helpers.PrepareAndRunTests("Dashboard", t, true)
}
