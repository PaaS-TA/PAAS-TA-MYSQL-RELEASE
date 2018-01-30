package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry-incubator/route-registrar/config"
	"github.com/fraenkel/candiedyaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

const (
	routeRegistrarPackage = "github.com/cloudfoundry-incubator/route-registrar/"
)

var (
	routeRegistrarBinPath string
	pidFile               string
	configFile            string
	scriptPath            string
	rootConfig            config.Config
	natsPort              int
)

func TestRouteRegistrar(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var _ = BeforeSuite(func() {
	var err error
	routeRegistrarBinPath, err = gexec.Build(routeRegistrarPackage, "-race")
	Ω(err).ShouldNot(HaveOccurred())

	tempDir, err := ioutil.TempDir(os.TempDir(), "route-registrar")
	Ω(err).NotTo(HaveOccurred())

	pidFile = filepath.Join(tempDir, "route-registrar.pid")

	configFile = filepath.Join(tempDir, "registrar_settings.yml")
	scriptPath = filepath.Join(tempDir,"check_node_validity.sh")
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func initConfig() {

	natsPort = 42222 + GinkgoParallelNode()

	messageBusServers := []config.MessageBusServer{
		config.MessageBusServer{
			Host:     fmt.Sprintf("127.0.0.1:%d", natsPort),
			User:     "nats",
			Password: "nats",
		},
	}

	healthCheckerConfig := &config.HealthCheckerConf{
		Name:     "riak-cs-cluster",
		Interval: 10,
	}

	rootConfig = config.Config{
		MessageBusServers: messageBusServers,
		ExternalHost:      "riakcs-vcap.me",
		ExternalIp:        "127.0.0.1",
		Port:              8080,
		HealthChecker:     healthCheckerConfig,
	}
}

func writeConfig() {
	fileToWrite, err := os.Create(configFile)
	Ω(err).ShouldNot(HaveOccurred())

	encoder := candiedyaml.NewEncoder(fileToWrite)
	err = encoder.Encode(rootConfig)
	Ω(err).ShouldNot(HaveOccurred())
}
