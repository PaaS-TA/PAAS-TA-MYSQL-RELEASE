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

var _ = Describe("RiakHealthChecker", func() {
	var logger lager.Logger
	Describe("Check", func() {
		var pidFilename string
		var riakAdminProgram string
		var scriptPath string

		BeforeEach(func() {
			logger = lagertest.NewTestLogger("RiakHealthChecker test")
			path, _ := os.Getwd()
			pidFilename = strings.Join([]string{path, "/../test_helpers/examplePidFile.pid"}, "")
			riakAdminProgram = strings.Join([]string{path, "/../test_helpers/riak-admin"}, "")
			scriptPath = strings.Join([]string{path, "/check_node_validity.sh"}, "")
		})

		It("returns true when the PID file exists and the node is a member of the cluster", func() {
			riakHealthChecker := NewRiakHealthChecker(pidFilename, riakAdminProgram, "1.2.3.4", logger, scriptPath)

			Expect(riakHealthChecker.Check()).To(BeTrue())
		})

		It("returns false when the PID file does not exist", func() {
			riakHealthChecker := NewRiakHealthChecker("/tmp/file-that-does-not-exist", riakAdminProgram, "1.2.3.4", logger, scriptPath)

			Expect(riakHealthChecker.Check()).To(BeFalse())
		})

		It("returns false when the PID file exists but the node is not present in the cluster", func() {
			riakHealthChecker := NewRiakHealthChecker(pidFilename, riakAdminProgram, "1.2.3.99", logger, scriptPath)

			Expect(riakHealthChecker.Check()).To(BeFalse())
		})

		It("returns false when the PID file exists and the node is present in the cluster but the node status is not 'valid'", func() {
			riakHealthChecker := NewRiakHealthChecker(pidFilename, riakAdminProgram, "1.2.3.5", logger, scriptPath)

			Expect(riakHealthChecker.Check()).To(BeFalse())
		})

		It("returns false when the riakAdminProgram returns an error", func() {
			riakHealthChecker := NewRiakHealthChecker(pidFilename, "admin_program_not_exist", "1.2.3.4", logger, scriptPath)

			Expect(riakHealthChecker.Check()).To(BeFalse())
		})
	})
})
