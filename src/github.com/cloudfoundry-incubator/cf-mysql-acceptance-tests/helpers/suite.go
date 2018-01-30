package helpers

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-test-helpers/services"
)

var TestConfig MysqlIntegrationConfig
var TestContext services.Context

func PrepareAndRunTests(packageName string, t *testing.T, withContext bool) {
	var err error
	TestConfig, err = LoadConfig()
	if err != nil {
		panic("Loading config: " + err.Error())
	}

	err = ValidateConfig(&TestConfig)
	if err != nil {
		panic("Validating config: " + err.Error())
	}

	TestContext = services.NewContext(TestConfig.Config, "MySQLATS")

	if withContext {
		BeforeEach(TestContext.Setup)
		AfterEach(TestContext.Teardown)
	}

	fmt.Printf("Plans: %#v\n", TestConfig.Plans)

	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("junit_%d.xml", ginkgoconfig.GinkgoConfig.ParallelNode))
	RunSpecsWithDefaultAndCustomReporters(t, fmt.Sprintf("P-MySQL Acceptance Tests -- %s", packageName), []Reporter{junitReporter})
}
