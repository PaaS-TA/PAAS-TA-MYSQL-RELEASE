package quota_test

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	. "github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"
)

const (
	// The quota enforcer sleeps for one second between iterations,
	// so sleeping for 20 seconds is sufficient for it have enforced all quotas
	quotaEnforcerSleepTime = 20 * time.Second
)

var _ = Describe("P-MySQL Service", func() {
	var sinatraPath = "../../assets/sinatra_app"

	assertAppIsRunning := func(appName string) {
		pingURI := helpers.TestConfig.AppURI(appName) + "/ping"
		fmt.Println("\n*** Checking that the app is responding at url: ", pingURI)

		runner.NewCmdRunner(runner.Curl(pingURI), helpers.TestContext.ShortTimeout()).WithAttempts(3).WithOutput("OK").Run()
	}

	Describe("Enforcing MySQL storage and connection quota", func() {
		var appName string
		var serviceInstanceName string
		var serviceURI string
		var plan helpers.Plan

		BeforeEach(func() {
			appName = RandomName()
			serviceInstanceName = RandomName()
			plan = helpers.TestConfig.Plans[0]

			serviceURI = fmt.Sprintf("%s/service/mysql/%s", helpers.TestConfig.AppURI(appName), serviceInstanceName)
			runner.NewCmdRunner(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), helpers.TestContext.LongTimeout()).Run()
		})

		JustBeforeEach(func() {
			fmt.Printf("Creating service with serviceName: %s, planName: %s, serviceInstanceName: %s\n", helpers.TestConfig.ServiceName, plan.Name, serviceInstanceName)
			runner.NewCmdRunner(Cf("create-service", helpers.TestConfig.ServiceName, plan.Name, serviceInstanceName), helpers.TestContext.LongTimeout()).Run()
			runner.NewCmdRunner(Cf("bind-service", appName, serviceInstanceName), helpers.TestContext.LongTimeout()).Run()
			runner.NewCmdRunner(Cf("start", appName), helpers.TestContext.LongTimeout()).Run()
			assertAppIsRunning(appName)
		})

		AfterEach(func() {
			runner.NewCmdRunner(Cf("unbind-service", appName, serviceInstanceName), helpers.TestContext.LongTimeout()).Run()
			runner.NewCmdRunner(Cf("delete-service", "-f", serviceInstanceName), helpers.TestContext.LongTimeout()).Run()
			runner.NewCmdRunner(Cf("delete", appName, "-f"), helpers.TestContext.LongTimeout()).Run()
		})

		ExceedLimit := func(maxStorageMb int) {
			writeUri := fmt.Sprintf("%s/write-bulk-data", serviceURI)

			fmt.Printf("\n*** Exceeding limit of %d\n", maxStorageMb)
			mbToWrite := 10
			loopIterations := (maxStorageMb / mbToWrite)

			for i := 0; i < loopIterations; i++ {
				curlCmd := runner.NewCmdRunner(runner.Curl("-v", "-d", strconv.Itoa(mbToWrite), writeUri), helpers.TestContext.ShortTimeout()).Run()
				Expect(curlCmd).To(Say("Database now contains"))
			}

			remainder := maxStorageMb % mbToWrite
			if remainder != 0 {
				curlCmd := runner.NewCmdRunner(runner.Curl("-v", "-d", strconv.Itoa(remainder), writeUri), helpers.TestContext.ShortTimeout()).Run()
				Expect(curlCmd).To(Say("Database now contains"))
			}

			// Write a little bit more to guarantee we are over quota
			// as opposed to being exactly at quota,
			// We are not interested in the output because we know we will be over quota.
			runner.NewCmdRunner(runner.Curl("-v", "-d", strconv.Itoa(1), writeUri), helpers.TestContext.ShortTimeout()).Run()
		}

		// We only need to validate the storage quota enforcer operates as expected over the first plan.
		// Especially important given the plan can be of any size, and we don't want to fill up large databases.
		It("enforces the storage quota for the plan", func() {
			uri := fmt.Sprintf("%s/mykey", serviceURI)
			deleteUri := fmt.Sprintf("%s/delete-bulk-data", serviceURI)
			firstValue := RandomName()[:20]
			secondValue := RandomName()[:20]

			fmt.Println("\n*** Proving we can write")
			curlCmd := runner.NewCmdRunner(runner.Curl("-d", firstValue, uri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say(firstValue))

			fmt.Println("\n*** Proving we can read")
			curlCmd = runner.NewCmdRunner(runner.Curl(uri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say(firstValue))

			ExceedLimit(plan.MaxStorageMb)

			fmt.Println("\n*** Sleeping to let quota enforcer run")
			time.Sleep(quotaEnforcerSleepTime)

			fmt.Println("\n*** Proving we cannot write (expect app to fail)")
			value := RandomName()[:20]
			curlCmd = runner.NewCmdRunner(runner.Curl("-d", value, uri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say("Error: (INSERT|UPDATE) command denied .* for table 'data_values'"))
			fmt.Println("Expected failure occured")

			fmt.Println("\n*** Proving we can read")
			curlCmd = runner.NewCmdRunner(runner.Curl(uri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say(firstValue))

			fmt.Println("\n*** Deleting below quota")
			curlCmd = runner.NewCmdRunner(runner.Curl("-d", "20", deleteUri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say("Database now contains"))

			fmt.Println("\n*** Sleeping to let quota enforcer run")
			time.Sleep(quotaEnforcerSleepTime)

			fmt.Println("\n*** Proving we can write")
			curlCmd = runner.NewCmdRunner(runner.Curl("-d", secondValue, uri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say(secondValue))

			fmt.Println("\n*** Proving we can read")
			curlCmd = runner.NewCmdRunner(runner.Curl(uri), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say(secondValue))
		})

		It("enforces the connection quota for the plan", func() {
			connectionsURI := fmt.Sprintf("%s/connections/mysql/%s/", helpers.TestConfig.AppURI(appName), serviceInstanceName)

			fmt.Println("\n*** Proving we can use the max num of connections")
			curlCmd := runner.NewCmdRunner(runner.Curl(connectionsURI+strconv.Itoa(plan.MaxUserConnections)), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say("success"))

			fmt.Println("\n*** Proving the connection quota is enforced")
			curlCmd = runner.NewCmdRunner(runner.Curl(connectionsURI+strconv.Itoa(plan.MaxUserConnections+1)), helpers.TestContext.ShortTimeout()).Run()
			Expect(curlCmd).To(Say("Error"), "Connection quota was not enforced. This may fail if proxies are behind a load balancer.")
		})

		Describe("Migrating a service instance between plans of different storage quota", func() {
			Context("when upgrading to a larger storage quota", func() {
				var newPlan helpers.Plan

				BeforeEach(func() {
					newPlan = helpers.TestConfig.Plans[1]
				})

				It("enforces the new quota", func() {
					uri := fmt.Sprintf("%s/mykey", serviceURI)
					ExceedLimit(plan.MaxStorageMb)

					fmt.Println("\n*** Sleeping to let quota enforcer run")
					time.Sleep(quotaEnforcerSleepTime)

					fmt.Println("\n*** Proving we cannot write (expect app to fail)")
					value := RandomName()[:20]
					curlCmd := runner.NewCmdRunner(runner.Curl("-d", value, uri), helpers.TestContext.ShortTimeout()).Run()
					Expect(curlCmd).To(Say("Error: (INSERT|UPDATE) command denied .* for table 'data_values'"))
					fmt.Println("Expected failure occured")

					fmt.Println("\n*** Upgrading service instance")
					cfCmd := runner.NewCmdRunner(Cf("update-service", serviceInstanceName, "-p", newPlan.Name), helpers.TestContext.LongTimeout()).Run()
					Expect(cfCmd).To(Say("OK"))

					fmt.Println("\n*** Sleeping to let quota enforcer run")
					time.Sleep(quotaEnforcerSleepTime)

					fmt.Println("\n*** Proving we can write")
					value = RandomName()[:20]
					curlCmd = runner.NewCmdRunner(runner.Curl("-d", value, uri), helpers.TestContext.ShortTimeout()).Run()
					Expect(curlCmd).To(Say(value))
				})
			})

			Context("when attempting to downgrade to a smaller storage quota", func() {
				var smallPlan helpers.Plan

				BeforeEach(func() {
					plan = helpers.TestConfig.Plans[1]
					smallPlan = helpers.TestConfig.Plans[0]
				})

				Context("when storage usage is over smaller quota", func() {
					It("disallows downgrade", func() {
						ExceedLimit(smallPlan.MaxStorageMb)

						fmt.Println("\n*** Sleeping to let quota enforcer run")
						time.Sleep(quotaEnforcerSleepTime)

						fmt.Println("\n*** Proving we can write")
						value := RandomName()[:20]
						uri := fmt.Sprintf("%s/mykey", serviceURI)
						curlCmd := runner.NewCmdRunner(runner.Curl("-d", value, uri), helpers.TestContext.ShortTimeout()).Run()
						Expect(curlCmd).To(Say(value))

						fmt.Println("\n*** Downgrading service instance (Expect failure)")
						cfCmd := runner.NewCmdRunner(Cf("update-service", serviceInstanceName, "-p", smallPlan.Name), helpers.TestContext.LongTimeout()).WithExitCode(1).Run()
						Expect(cfCmd).To(Say("Service broker error"))
						fmt.Println("Expected failure occured")
					})
				})

				Context("when storage usage is under smaller quota", func() {
					It("allows downgrade", func() {
						ExceedLimit(0)

						fmt.Println("\n*** Sleeping to let quota enforcer run")
						time.Sleep(quotaEnforcerSleepTime)

						fmt.Println("\n*** Proving we can write")
						value := RandomName()[:20]
						uri := fmt.Sprintf("%s/mykey", serviceURI)
						curlCmd := runner.NewCmdRunner(runner.Curl("-d", value, uri), helpers.TestContext.ShortTimeout()).Run()
						Expect(curlCmd).To(Say(value))

						fmt.Println("\n*** Downgrading service instance")
						cfCmd := runner.NewCmdRunner(Cf("update-service", serviceInstanceName, "-p", smallPlan.Name), helpers.TestContext.LongTimeout()).Run()
						Expect(cfCmd).To(Say("OK"))

						fmt.Println("\n*** Sleeping to let quota enforcer run")
						time.Sleep(quotaEnforcerSleepTime)

						fmt.Println("\n*** Proving we can write")
						value = RandomName()[:20]
						curlCmd = runner.NewCmdRunner(runner.Curl("-d", value, uri), helpers.TestContext.ShortTimeout()).Run()
						Expect(curlCmd).To(Say(value))
					})
				})
			})
		})
	})
})
