package broker_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	cfhelpers "github.com/cloudfoundry-incubator/cf-test-helpers/helpers"
)

var _ = Describe("P-MySQL Service broker", func() {
	It("Denies access to the catalog endpoint when credentials are not provided", func() {
		uri := fmt.Sprintf("%s://%s/v2/catalog", helpers.TestConfig.BrokerProtocol, helpers.TestConfig.BrokerHost)

		fmt.Printf("\n*** Curling url: %s\n", uri)
		curlCmd := cfhelpers.Curl(helpers.TestConfig.CFConfig, uri).Wait(helpers.TestContext.ShortTimeout())
		Expect(curlCmd).To(Say("HTTP Basic: Access denied."))
		fmt.Println("Expected failure occured")
	})
})
