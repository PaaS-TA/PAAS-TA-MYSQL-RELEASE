package failover_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/dsl"

	. "github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/runner"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/partition"
)

const (
	firstKey    = "mykey"
	firstValue  = "myvalue"
	secondKey   = "mysecondkey"
	secondValue = "mysecondvalue"
	planName    = "100mb"

	sinatraPath = "../../assets/sinatra_app"

	// The route takes 2 minutes to prune routes; wait 2.5 minutes to ensure
	// we only route to the remaining broker.
	routeRegistrarPruneSleepDuration = 150 * time.Second
)

func assertAppIsRunning(appName string) {
	pingURI := helpers.TestConfig.AppURI(appName) + "/ping"
	runner.NewCmdRunner(runner.Curl(pingURI), helpers.TestContext.ShortTimeout()).WithOutput("OK").Run()
}

func assertWriteToDB(key, value, uri string) {
	curlURI := fmt.Sprintf("%s/%s", uri, key)
	runner.NewCmdRunner(runner.Curl("-d", value, curlURI), helpers.TestContext.ShortTimeout()).WithOutput(value).Run()
}

func assertReadFromDB(key, value, uri string) {
	curlURI := fmt.Sprintf("%s/%s", uri, key)
	runner.NewCmdRunner(runner.Curl(curlURI), helpers.TestContext.ShortTimeout()).WithOutput(value).Run()
}

var _ = Feature("CF MySQL Failover", func() {
	var appName string
	var broker0SshTunnel, broker1SshTunnel string

	BeforeEach(func() {
		Expect(helpers.TestConfig.MysqlNodes).NotTo(BeNil())
		Expect(len(helpers.TestConfig.MysqlNodes)).To(BeNumerically(">=", 1))

		Expect(helpers.TestConfig.Brokers).NotTo(BeNil())
		Expect(len(helpers.TestConfig.Brokers)).To(BeNumerically(">=", 2))

		broker0SshTunnel = helpers.TestConfig.Brokers[0].SshTunnel
		broker1SshTunnel = helpers.TestConfig.Brokers[1].SshTunnel

		// Remove broker partitions in case previous test did not cleanup correctly
		partition.Off(broker0SshTunnel)
		partition.Off(broker1SshTunnel)

		appName = generator.RandomName()

		Step("Push an app", func() {
			runner.NewCmdRunner(Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-no-start"), helpers.TestContext.LongTimeout()).Run()
		})
	})

	AfterEach(func() {
		partition.Off(broker0SshTunnel)
		partition.Off(broker1SshTunnel)
		// TODO: Reintroduce the mariadb node once #81974864 is complete.
		// partition.Off(IntegrationConfig.MysqlNodes[0].SshTunnel)
	})

	Scenario("write/read data before and after a partition of mysql node and then broker", func() {
		instanceCount := 3
		serviceInstanceName := make([]string, instanceCount)
		instanceURI := make([]string, instanceCount)

		for i := 0; i < instanceCount; i++ {
			serviceInstanceName[i] = generator.RandomName()
			instanceURI[i] = helpers.TestConfig.AppURI(appName) + "/service/mysql/" + serviceInstanceName[i]
		}

		fmt.Println("MYSQL NODE FAILOVER")

		Step("Creating service instance[0]", func() {
			runner.NewCmdRunner(Cf("create-service", helpers.TestConfig.ServiceName, planName, serviceInstanceName[0]), helpers.TestContext.LongTimeout()).Run()
		})

		Step("Binding app to instance[0]", func() {
			runner.NewCmdRunner(Cf("bind-service", appName, serviceInstanceName[0]), helpers.TestContext.LongTimeout()).Run()
		})

		Step("Start app for the first time", func() {
			runner.NewCmdRunner(Cf("start", appName), helpers.TestContext.LongTimeout()).Run()
			assertAppIsRunning(appName)
		})

		Step("Write a key-value pair to instance[0]", func() {
			assertWriteToDB(firstKey, firstValue, instanceURI[0])
		})

		Step("Read value from instance[0]", func() {
			assertReadFromDB(firstKey, firstValue, instanceURI[0])
		})

		Step("Take down mysql node", func() {
			partition.On(
				helpers.TestConfig.MysqlNodes[0].SshTunnel,
				helpers.TestConfig.MysqlNodes[0].Ip,
			)
		})

		Step("Sleep to allow proxy time to fail-over", func() {
			time.Sleep(helpers.TestContext.ShortTimeout())
		})

		Step("Write a second key-value pair to instance[0]", func() {
			assertWriteToDB(secondKey, secondValue, instanceURI[0])
		})

		Step("Read both values from instance[0]", func() {
			assertReadFromDB(firstKey, firstValue, instanceURI[0])
			assertReadFromDB(secondKey, secondValue, instanceURI[0])
		})

		// Perform broker failover
		fmt.Println("BROKER FAILOVER")

		Step("Take down first broker instance", func() {
			partition.On(broker0SshTunnel, helpers.TestConfig.Brokers[0].Ip)
		})

		Step("Sleep to let route-registrar prune broker", func() {
			time.Sleep(routeRegistrarPruneSleepDuration)
		})

		Step("Creating service instance[1]", func() {
			runner.NewCmdRunner(Cf("create-service", helpers.TestConfig.ServiceName, planName, serviceInstanceName[1]), helpers.TestContext.LongTimeout()).Run()
		})

		Step("Binding app to instance[1]", func() {
			runner.NewCmdRunner(Cf("bind-service", appName, serviceInstanceName[1]), helpers.TestContext.LongTimeout()).Run()
		})

		Step("Restart App to receive updated service instance info", func() {
			runner.NewCmdRunner(Cf("restart", appName), helpers.TestContext.LongTimeout()).Run()
			assertAppIsRunning(appName)
		})

		Step("Write a key-value pair to DB", func() {
			assertWriteToDB(firstKey, firstValue, instanceURI[1])
		})

		Step("Read value from DB", func() {
			assertReadFromDB(firstKey, firstValue, instanceURI[1])
		})

		Step("Bring back first broker instance", func() {
			partition.Off(broker0SshTunnel)
		})

		Step("Take down second broker instance", func() {
			partition.On(broker1SshTunnel, helpers.TestConfig.Brokers[1].Ip)
		})

		Step("Sleep to let route-registrar prune broker", func() {
			time.Sleep(routeRegistrarPruneSleepDuration)
		})

		Step("Creating service instance[2]", func() {
			runner.NewCmdRunner(Cf("create-service", helpers.TestConfig.ServiceName, planName, serviceInstanceName[2]), helpers.TestContext.LongTimeout()).Run()
		})

		Step("Binding app to instance[2]", func() {
			runner.NewCmdRunner(Cf("bind-service", appName, serviceInstanceName[2]), helpers.TestContext.LongTimeout()).Run()
		})

		Step("Restart App to receive updated service instance info", func() {
			runner.NewCmdRunner(Cf("restart", appName), helpers.TestContext.LongTimeout()).Run()
			assertAppIsRunning(appName)
		})

		Step("Write a second key-value pair to DB", func() {
			assertWriteToDB(secondKey, secondValue, instanceURI[2])
		})

		Step("Read values from both bindings' DBs", func() {
			assertReadFromDB(firstKey, firstValue, instanceURI[1])
			assertReadFromDB(secondKey, secondValue, instanceURI[2])
		})

		Step("Bring back second broker instance", func() {
			partition.Off(broker1SshTunnel)
		})
	})
})
