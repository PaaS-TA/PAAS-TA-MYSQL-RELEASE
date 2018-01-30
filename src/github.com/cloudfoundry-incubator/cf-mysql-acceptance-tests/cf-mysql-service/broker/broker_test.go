package broker_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"
)

var _ = Describe("P-MySQL Service broker", func() {
	It("Registers a route", func() {
		uri := fmt.Sprintf("http://%s/v2/catalog", helpers.TestConfig.BrokerHost)

		fmt.Printf("\n*** Curling url: %s\n", uri)
		curlCmd := runner.NewCmdRunner(runner.Curl(uri), helpers.TestContext.ShortTimeout()).Run()
		Expect(curlCmd).To(Say("HTTP Basic: Access denied."))
		fmt.Println("Expected failure occured")
	})
})
