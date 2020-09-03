package quota_test

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
)

const (
	// The quota enforcer sleeps for one second between iterations,
	// so sleeping for 20 seconds is sufficient for it have enforced all quotas
	quotaEnforcerSleepTime = 20 * time.Second
)

var _ = Describe("P-MySQL Service", func() {
	var sinatraPath = "../../assets/sinatra_app"

	Describe("Enforcing MySQL storage and connection quota", func() {
		var appName string
		var serviceInstanceName string
		var plan helpers.Plan
		var appClient helpers.SinatraAppClient

		BeforeEach(func() {
			appName = generator.PrefixedRandomName("quota", "app")
			serviceInstanceName = generator.PrefixedRandomName("quota", "instance")
			plan = helpers.TestConfig.Plans[0]
			appClient = helpers.NewSinatraAppClient(helpers.TestConfig.AppURI(appName), serviceInstanceName, helpers.TestConfig.CFConfig.SkipSSLValidation)

			Expect(cf.Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-b", "ruby_buildpack", "--no-start").Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
		})

		JustBeforeEach(func() {
			fmt.Printf("Creating service with serviceName: %s, planName: %s, serviceInstanceName: %s\n", helpers.TestConfig.ServiceName, plan.Name, serviceInstanceName)
			Expect(cf.Cf("create-service", helpers.TestConfig.ServiceName, plan.Name, serviceInstanceName).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
			Expect(cf.Cf("bind-service", appName, serviceInstanceName).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
			Expect(cf.Cf("start", appName).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
			err := appClient.Ping()
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(cf.Cf("unbind-service", appName, serviceInstanceName).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
			Expect(cf.Cf("delete-service", "-f", serviceInstanceName).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
			Expect(cf.Cf("delete", appName, "-f").Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
		})

		ExceedLimit := func(maxStorageMb int) {
			fmt.Printf("\n*** Exceeding limit of %d\n", maxStorageMb)
			mbToWrite := 10
			loopIterations := (maxStorageMb / mbToWrite)

			for i := 0; i < loopIterations; i++ {
				msg, err := appClient.WriteBulkData(strconv.Itoa(mbToWrite))
				Expect(err).NotTo(HaveOccurred())
				Expect(msg).To(ContainSubstring("Database now contains"))
			}

			remainder := maxStorageMb % mbToWrite
			if remainder != 0 {
				msg, err := appClient.WriteBulkData(strconv.Itoa(remainder))
				Expect(err).NotTo(HaveOccurred())
				Expect(msg).To(ContainSubstring("Database now contains"))
			}

			// Write a little bit more to guarantee we are over quota
			// as opposed to being exactly at quota,
			// We are not interested in the output because we know we will be over quota.
			appClient.WriteBulkData(strconv.Itoa(1))
		}

		// We only need to validate the storage quota enforcer operates as expected over the first plan.
		// Especially important given the plan can be of any size, and we don't want to fill up large databases.
		It("enforces the storage quota for the plan", func() {
			firstValue := generator.PrefixedRandomName("", "")[:20]
			secondValue := generator.PrefixedRandomName("", "")[:20]

			fmt.Println("\n*** Proving we can write")
			msg, err := appClient.Set("mykey", firstValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(ContainSubstring(firstValue))

			fmt.Println("\n*** Proving we can read")
			msg, err = appClient.Get("mykey")
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(ContainSubstring(firstValue))

			ExceedLimit(plan.MaxStorageMb)

			fmt.Println("\n*** Sleeping to let quota enforcer run")
			time.Sleep(quotaEnforcerSleepTime)

			fmt.Println("\n*** Proving we cannot write (expect app to fail)")
			value := generator.PrefixedRandomName("", "")[:20]
			_, err = appClient.Set("mykey", value)
			Expect(err).To(MatchError(MatchRegexp("Error: (INSERT|UPDATE) command denied .* for table 'data_values'")))

			fmt.Println("Expected failure occured")

			fmt.Println("\n*** Proving we can read")
			msg, err = appClient.Get("mykey")
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(ContainSubstring(firstValue))

			fmt.Println("\n*** Deleting below quota")
			msg, err = appClient.DeleteBulkData("20")
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(ContainSubstring("Database now contains"))

			fmt.Println("\n*** Sleeping to let quota enforcer run")
			time.Sleep(quotaEnforcerSleepTime)

			fmt.Println("\n*** Proving we can write")
			msg, err = appClient.Set("mykey", secondValue)
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(ContainSubstring(secondValue))

			fmt.Println("\n*** Proving we can read")
			msg, err = appClient.Get("mykey")
			Expect(err).NotTo(HaveOccurred())
			Expect(msg).To(ContainSubstring(secondValue))
		})

		// TODO: Enable this test once we complete the proxy sync epic.
		// Dijon sends these connections through a load balancer, routing connections to both proxies that sometimes
		// choose different nodes. max_user_connections are not replicated across both nodes yet, so the test fails at
		// 41 (max_user_connections * 2) connections instead of 21.
		//
		// It("enforces the connection quota for the plan", func() {
		// 	connectionsURI := fmt.Sprintf("%s/connections/mysql/%s/", helpers.TestConfig.AppURI(appName), serviceInstanceName)

		// 	fmt.Println("\n*** Proving we can use the max num of connections")
		// 	curlCmd := runner.NewCmdRunner(runner.Curl("-k", connectionsURI+strconv.Itoa(plan.MaxUserConnections)), helpers.TestContext.ShortTimeout()).Run()
		// 	Expect(curlCmd).To(Say("success"))

		// 	fmt.Println("\n*** Proving the connection quota is enforced")
		// 	curlCmd = runner.NewCmdRunner(runner.Curl("-k", connectionsURI+strconv.Itoa(plan.MaxUserConnections+1)), helpers.TestContext.ShortTimeout()).Run()
		// 	Expect(curlCmd).To(Say("Error"), "Connection quota was not enforced. This may fail if proxies are behind a load balancer.")
		// })

		Describe("Migrating a service instance between plans of different storage quota", func() {
			Context("when upgrading to a larger storage quota", func() {
				var newPlan helpers.Plan

				BeforeEach(func() {
					newPlan = helpers.TestConfig.Plans[1]
				})

				It("enforces the new quota", func() {
					ExceedLimit(plan.MaxStorageMb)

					fmt.Println("\n*** Sleeping to let quota enforcer run")
					time.Sleep(quotaEnforcerSleepTime)

					fmt.Println("\n*** Proving we cannot write (expect app to fail)")
					value := generator.PrefixedRandomName("", "")[:20]
					_, err := appClient.Set("mykey", value)
					Expect(err).To(MatchError(MatchRegexp("Error: (INSERT|UPDATE) command denied .* for table 'data_values'")))
					fmt.Println("Expected failure occured")

					fmt.Println("\n*** Upgrading service instance")
					Expect(cf.Cf("update-service", serviceInstanceName, "-p", newPlan.Name).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))

					fmt.Println("\n*** Sleeping to let quota enforcer run")
					time.Sleep(quotaEnforcerSleepTime)

					fmt.Println("\n*** Proving we can write")
					value = generator.PrefixedRandomName("", "")[:20]
					//curlCmd = runner.NewCmdRunner(runner.Curl("-k", "-d", value, uri), helpers.TestContext.ShortTimeout()).Run()
					msg, err := appClient.Set("mykey", value)
					Expect(err).NotTo(HaveOccurred())
					Expect(msg).To(ContainSubstring(value))
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
						value := generator.PrefixedRandomName("", "")[:20]

						msg, err := appClient.Set("mykey", value)
						Expect(err).NotTo(HaveOccurred())
						Expect(msg).To(ContainSubstring(value))

						fmt.Println("\n*** Downgrading service instance (Expect failure)")
						cfCmd := cf.Cf("update-service", serviceInstanceName, "-p", smallPlan.Name).Wait(helpers.TestContext.LongTimeout())
						Expect(cfCmd).To(Exit(1))
						Expect(cfCmd.Out).To(Say("Service broker error"))

						fmt.Println("Expected failure occured")
					})
				})

				Context("when storage usage is under smaller quota", func() {
					It("allows downgrade", func() {
						ExceedLimit(0)

						fmt.Println("\n*** Sleeping to let quota enforcer run")
						time.Sleep(quotaEnforcerSleepTime)

						fmt.Println("\n*** Proving we can write")
						value := generator.PrefixedRandomName("", "")[:20]
						msg, err := appClient.Set("mykey", value)
						Expect(err).NotTo(HaveOccurred())
						Expect(msg).To(ContainSubstring(value))

						fmt.Println("\n*** Downgrading service instance")
						Expect(cf.Cf("update-service", serviceInstanceName, "-p", smallPlan.Name).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))

						fmt.Println("\n*** Sleeping to let quota enforcer run")
						time.Sleep(quotaEnforcerSleepTime)

						fmt.Println("\n*** Proving we can write")
						value = generator.PrefixedRandomName("", "")[:20]
						msg, err = appClient.Set("mykey", value)
						Expect(err).NotTo(HaveOccurred())
						Expect(msg).To(ContainSubstring(value))
					})
				})
			})
		})
	})
})
