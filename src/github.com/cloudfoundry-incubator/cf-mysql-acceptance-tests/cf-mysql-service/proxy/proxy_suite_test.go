package proxy_test

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"

	. "github.com/onsi/ginkgo"
)

func TestService(t *testing.T) {
	helpers.PrepareAndRunTests("Proxy", t, false)
}

var _ = BeforeSuite(func() {
	if helpers.TestConfig.Proxy.SkipSSLValidation {
		http.DefaultClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: helpers.TestConfig.Proxy.SkipSSLValidation,
				},
			},
		}
	}
})
