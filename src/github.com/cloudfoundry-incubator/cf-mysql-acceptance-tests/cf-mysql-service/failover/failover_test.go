package failover_test

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"

	boshuaa "github.com/cloudfoundry/bosh-cli/uaa"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	"strings"
)

const (
	firstKey    = "mykey"
	firstValue  = "myvalue"
	secondKey   = "mysecondkey"
	secondValue = "mysecondvalue"
	planName    = "10mb"

	sinatraPath = "../../assets/sinatra_app"
)

func buildUAA() (boshuaa.UAA, error) {
	logger := boshlog.NewLogger(boshlog.LevelError)
	factory := boshuaa.NewFactory(logger)

	// Build a UAA config from a URL.
	// HTTPS is required and certificates are always verified.
	config, err := boshuaa.NewConfigFromURL(fmt.Sprintf("https://%s:8443", helpers.TestConfig.BOSH.URL))
	if err != nil {
		return nil, err
	}

	// Set client credentials for authentication.
	// Machine level access should typically use a client instead of a particular user.
	config.Client = helpers.TestConfig.BOSH.Client
	config.ClientSecret = helpers.TestConfig.BOSH.ClientSecret

	// Configure trusted CA certificates.
	// If nothing is provided default system certificates are used.
	config.CACert = helpers.TestConfig.BOSH.CACert

	return factory.New(config)
}

func buildDirector(uaa boshuaa.UAA) (boshdir.Director, error) {
	logger := boshlog.NewLogger(boshlog.LevelError)
	factory := boshdir.NewFactory(logger)

	// Build a Director config from address-like string.
	// HTTPS is required and certificates are always verified.
	config, err := boshdir.NewConfigFromURL(helpers.TestConfig.BOSH.URL)
	if err != nil {
		return nil, err
	}

	// Configure custom trusted CA certificates.
	// If nothing is provided default system certificates are used.
	config.CACert = helpers.TestConfig.BOSH.CACert

	// Allow Director to fetch UAA tokens when necessary.
	config.TokenFunc = boshuaa.NewClientTokenSession(uaa).TokenFunc

	return factory.New(config, boshdir.NewNoopTaskReporter(), boshdir.NewNoopFileReporter())
}

func deleteMysqlVM(host string) error {
	uaa, err := buildUAA()
	if err != nil {
		return err
	}

	director, err := buildDirector(uaa)
	if err != nil {
		return err
	}

	deployment, err := director.FindDeployment("cf-mysql")
	if err != nil {
		return err
	}

	instances, err := deployment.Instances()
	if err != nil {
		return err
	}

	var vmcid string
	for _, instance := range instances {
		if instance.Group == "mysql" {
			hostArray := strings.Split(host, ".")
			if instance.IPs[0] == host || (len(hostArray) > 0 && hostArray[0] == instance.ID) {
				vmcid = instance.VMID
				break
			}
		}
	}

	if vmcid == "" {
		return fmt.Errorf("no vm found with %s", host)
	}

	return deployment.DeleteVM(vmcid)
}

func activeProxyBackend() (string, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v0/cluster", helpers.TestConfig.Proxy.DashboardUrls[0]), nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(helpers.TestConfig.Proxy.APIUsername, helpers.TestConfig.Proxy.APIPassword)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var cluster struct {
		ActiveBackend struct {
			Host string `json:"host"`
		} `json:"activeBackend`
	}

	if err := json.Unmarshal(body, &cluster); err != nil {
		return "", err
	}

	return cluster.ActiveBackend.Host, nil
}

var _ = Describe("CF MySQL Failover", func() {
	It("write/read data before and after a partition of mysql node", func() {
		var oldBackend string

		serviceInstanceName := generator.PrefixedRandomName("failover", "instance")
		appName := generator.PrefixedRandomName("failover", "app")

		var appClient = helpers.NewSinatraAppClient(helpers.TestConfig.AppURI(appName), serviceInstanceName, helpers.TestConfig.CFConfig.SkipSSLValidation)

		Expect(cf.Cf("push", appName, "-m", "256M", "-p", sinatraPath, "-b", "ruby_buildpack", "--no-start").
			Wait(helpers.TestContext.LongTimeout())).
			To(Exit(0))

		Expect(cf.Cf("create-service", helpers.TestConfig.ServiceName, planName, serviceInstanceName).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
		Expect(cf.Cf("bind-service", appName, serviceInstanceName).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
		Expect(cf.Cf("start", appName).Wait(helpers.TestContext.LongTimeout())).To(Exit(0))
		err := appClient.Ping()
		Expect(err).NotTo(HaveOccurred())

		msg, err := appClient.Set(firstKey, firstValue)
		Expect(msg).To(ContainSubstring(firstValue))
		Expect(err).NotTo(HaveOccurred())

		msg, err = appClient.Get(firstKey)
		Expect(msg).To(ContainSubstring(firstValue))
		Expect(err).NotTo(HaveOccurred())

		By("querying the proxy for the current mysql backend", func() {
			var err error

			oldBackend, err = activeProxyBackend()
			Expect(err).NotTo(HaveOccurred())
		})

		By("Take down the active mysql node", func() {
			err := deleteMysqlVM(oldBackend)
			Expect(err).NotTo(HaveOccurred())

		})

		By("poll the proxy for a backend change", func() {
			Eventually(func() bool {
				backend, err := activeProxyBackend()
				Expect(err).NotTo(HaveOccurred())

				return backend != oldBackend
			}, 5*time.Minute, 20*time.Second).Should(BeTrue())
		})

		msg, err = appClient.Set(secondKey, secondValue)
		Expect(msg).To(ContainSubstring(secondValue))
		Expect(err).NotTo(HaveOccurred())

		msg, err = appClient.Get(firstKey)
		Expect(msg).To(ContainSubstring(firstValue))
		Expect(err).NotTo(HaveOccurred())

		msg, err = appClient.Get(secondKey)
		Expect(msg).To(ContainSubstring(secondValue))
		Expect(err).NotTo(HaveOccurred())
	})
})
