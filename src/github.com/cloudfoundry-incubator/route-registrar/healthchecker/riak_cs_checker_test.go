package healthchecker_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/route-registrar/healthchecker"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("RiakCSHealthChecker", func() {
	var logger lager.Logger
	Describe("Check", func() {
		var pidFilename string
		var riakAdminProgram string

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("RiakCSHealthChecker test")
			path, _ := os.Getwd()
			pidFilename = strings.Join([]string{path, "/../test_helpers/examplePidFile.pid"}, "")
			riakAdminProgram = strings.Join([]string{path, "/../test_helpers/riak-admin"}, "")
		})

		It("returns true when the PID file exists", func() {
			riakCsHealthChecker := NewRiakCSHealthChecker(pidFilename, logger)

			Expect(riakCsHealthChecker.Check()).To(BeTrue())
		})

		It("returns false when the PID file does not exist", func() {
			riakCsHealthChecker := NewRiakCSHealthChecker("/tmp/file-that-does-not-exist", logger)

			Expect(riakCsHealthChecker.Check()).To(BeFalse())
		})
	})
})
