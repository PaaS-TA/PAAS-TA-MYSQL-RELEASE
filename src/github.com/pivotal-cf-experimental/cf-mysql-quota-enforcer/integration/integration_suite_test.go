package enforcer_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/pivotal-cf-experimental/cf-mysql-quota-enforcer/config"
	"github.com/pivotal-cf-experimental/cf-mysql-quota-enforcer/database"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pivotal-cf-experimental/service-config"
)

var brokerDBName string
var rootConfig config.Config
var binaryPath string

var tempDir string
var configPath string

func TestEnforcer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Enforcer Suite")
}

func newRootDatabaseConfig(dbName string) config.Config {
	serviceConfig := service_config.New()

	var dbConfig config.Config
	err := serviceConfig.Read(&dbConfig)
	Expect(err).ToNot(HaveOccurred())

	dbConfig.DBName = dbName

	err = dbConfig.Validate()
	Expect(err).ToNot(HaveOccurred())

	return dbConfig
}

var _ = BeforeSuite(func() {
	initConfig := newRootDatabaseConfig("")

	brokerDBName = uuidWithUnderscores("db")
	rootConfig = newRootDatabaseConfig(brokerDBName)

	initDB, err := database.NewConnection(initConfig)
	Expect(err).ToNot(HaveOccurred())
	defer initDB.Close()

	_, err = initDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", brokerDBName))
	Expect(err).ToNot(HaveOccurred())

	db, err := database.NewConnection(rootConfig)
	Expect(err).ToNot(HaveOccurred())
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS service_instances (
    id int(11) NOT NULL AUTO_INCREMENT,
    guid varchar(255),
    plan_guid varchar(255),
    max_storage_mb int(11) NOT NULL DEFAULT '0',
    db_name varchar(255),
    PRIMARY KEY (id)
	)`)
	Expect(err).ToNot(HaveOccurred())

	binaryPath, err = gexec.Build("github.com/pivotal-cf-experimental/cf-mysql-quota-enforcer", "-race")
	Expect(err).ToNot(HaveOccurred())

	_, err = os.Stat(binaryPath)
	if err != nil {
		Expect(os.IsExist(err)).To(BeTrue())
	}

	tempDir, err = ioutil.TempDir(os.TempDir(), "quota-enforcer-integration-test")
	Expect(err).NotTo(HaveOccurred())

	configPath = filepath.Join(tempDir, "quotaEnforcerConfig.yml")
	writeConfig()
})

var _ = AfterSuite(func() {

	// We don't need to handle an error cleaning up the tempDir
	_ = os.RemoveAll(tempDir)

	gexec.CleanupBuildArtifacts()

	_, err := os.Stat(binaryPath)
	if err != nil {
		Expect(os.IsExist(err)).To(BeFalse())
	}

	var emptyConfig config.Config
	if rootConfig != emptyConfig {
		db, err := database.NewConnection(rootConfig)
		Expect(err).ToNot(HaveOccurred())
		defer db.Close()

		_, err = db.Exec("DROP TABLE IF EXISTS service_instances")
		Expect(err).ToNot(HaveOccurred())

		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", brokerDBName))
		Expect(err).ToNot(HaveOccurred())
	}
})

func startEnforcerWithFlags(flags ...string) *gexec.Session {

	flags = append(
		flags,
		fmt.Sprintf("-configPath=%s", configPath),
		"-logLevel=debug",
	)

	command := exec.Command(
		binaryPath,
		flags...,
	)

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	return session
}

func runEnforcerContinuously(flags ...string) *gexec.Session {
	session := startEnforcerWithFlags(flags...)
	Eventually(session.Out).Should(gbytes.Say("Running continuously"))
	return session
}

func runEnforcerOnce() {
	session := startEnforcerWithFlags("-runOnce")

	Eventually(session.Out).Should(gbytes.Say("Running once"))
	// Wait for the process to finish naturally.
	// This should not take a long time
	session.Wait(5 * time.Second)
	Expect(session.ExitCode()).To(Equal(0), string(session.Err.Contents()))
}

func writeConfig() {
	fileToWrite, err := os.Create(configPath)
	Expect(err).ToNot(HaveOccurred())

	bytes, err := json.MarshalIndent(rootConfig, "", "  ")
	Expect(err).ToNot(HaveOccurred())

	_, err = fileToWrite.Write(bytes)
	Expect(err).ToNot(HaveOccurred())
}

func uuidWithUnderscores(prefix string) string {
	id, err := uuid.NewV4()
	Expect(err).ToNot(HaveOccurred())
	idString := fmt.Sprintf("%s_%s", prefix, id.String())
	return strings.Replace(idString, "-", "_", -1)
}
