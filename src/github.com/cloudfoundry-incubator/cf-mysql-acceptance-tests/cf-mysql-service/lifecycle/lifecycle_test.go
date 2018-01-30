package lifecycle_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"
)

var _ = Describe("P-MySQL Lifecycle Tests", func() {
	var sinatraPath = "../../assets/sinatra_app"

	assertAppIsRunning := func(appName string) {
		pingURI := helpers.TestConfig.AppURI(appName) + "/ping"
		fmt.Println("\n*** Checking that the app is responding at url: ", pingURI)

		runner.NewCmdRunner(runner.Curl(pingURI), helpers.TestContext.ShortTimeout()).WithAttempts(3).WithOutput("OK").Run()
	}

	It("Allows users to create, bind, write to, read from, unbind, and destroy a service instance for the each plan", func() {
		for _, plan := range helpers.TestConfig.Plans {
			appName := RandomName()
			pushCmd := runner.NewCmdRunner(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), helpers.TestContext.LongTimeout()).Run()
			Expect(pushCmd).To(Say("OK"))

			serviceInstanceName := RandomName()
			uri := fmt.Sprintf("%s/service/mysql/%s/mykey", helpers.TestConfig.AppURI(appName), serviceInstanceName)

			runner.NewCmdRunner(Cf("create-service", helpers.TestConfig.ServiceName, plan.Name, serviceInstanceName), helpers.TestContext.LongTimeout()).Run()

			runner.NewCmdRunner(Cf("bind-service", appName, serviceInstanceName), helpers.TestContext.LongTimeout()).Run()
			runner.NewCmdRunner(Cf("start", appName), helpers.TestContext.LongTimeout()).Run()
			assertAppIsRunning(appName)

			fmt.Printf("\n*** Posting to url: %s\n", uri)
			curlCmd := runner.NewCmdRunner(runner.Curl("-d", "myvalue", uri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say("myvalue"))

			fmt.Printf("\n*** Curling url: %s\n", uri)
			curlCmd = runner.NewCmdRunner(runner.Curl(uri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say("myvalue"))

			runner.NewCmdRunner(Cf("unbind-service", appName, serviceInstanceName), helpers.TestContext.LongTimeout()).Run()
			runner.NewCmdRunner(Cf("delete-service", "-f", serviceInstanceName), helpers.TestContext.LongTimeout()).Run()

			runner.NewCmdRunner(Cf("delete", appName, "-f"), helpers.TestContext.LongTimeout()).Run()
		}
	})
})
