package proxy_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
)

var _ = Describe("P-MySQL Proxy", func() {

	It("prompts for Basic Auth creds when they aren't provided", func() {
		for _, url := range helpers.TestConfig.Proxy.DashboardUrls {
			resp, err := http.Get(url)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		}
	})

	It("does not accept bad Basic Auth creds", func() {
		for _, url := range helpers.TestConfig.Proxy.DashboardUrls {
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth("bad_username", "bad_password")
			resp, err := http.DefaultClient.Do(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		}
	})

	It("accepts valid Basic Auth creds", func() {
		for _, url := range helpers.TestConfig.Proxy.DashboardUrls {
			req, err := http.NewRequest("GET", url, nil)
			req.SetBasicAuth(
				helpers.TestConfig.Proxy.APIUsername,
				helpers.TestConfig.Proxy.APIPassword,
			)
			resp, err := http.DefaultClient.Do(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}
	})
})
