package helpers

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	ginkgoconfig "github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
)

var TestConfig MysqlIntegrationConfig
var TestContext *workflowhelpers.ReproducibleTestSuiteSetup

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

	if withContext {
		BeforeEach(func() {
			TestContext = workflowhelpers.NewTestSuiteSetup(TestConfig.CFConfig)
			TestContext.Setup()
		})

		AfterEach(func() {
			TestContext.Teardown()
		})
	}

	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("junit_%d.xml", ginkgoconfig.GinkgoConfig.ParallelNode))
	RunSpecsWithDefaultAndCustomReporters(t, fmt.Sprintf("P-MySQL Acceptance Tests -- %s", packageName), []Reporter{junitReporter})
}
