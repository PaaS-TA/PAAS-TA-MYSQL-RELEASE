package main_test

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"bytes"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Configuration Generator Command", func() {
	var (
		tmpDir       string
		manifestPath string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "manifest-dir")
		Expect(err).NotTo(HaveOccurred())

		manifestPath = filepath.Join(tmpDir, "manifest.yml")
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	It("turns a manifest into the correct integration configuration", func() {
		config := `{
				"api": "https://api.bosh-lite.com",
				"apps_domain": "example.com",
				"admin_user": "admin",
				"admin_password": "foobar",
				"broker_host": "p-mysql.bosh-lite.com",
				"service_name": "p-mysql",
				"create_permissive_security_group": false,
				"org_name": "system",
				"plans" : [
					{
						"plan_name": "10mb",
						"max_user_connections": 20,
						"max_storage_mb": 10,
						"private": true
					},
					{
						"plan_name": "100mb",
						"max_user_connections": 40,
						"max_storage_mb": 100
					}
				],
				"skip_ssl_validation": true,
				"proxy": {
					"dashboard_urls": [
						"https://proxy-0-p-mysql.bosh-lite.com",
						"https://proxy-1-p-mysql.bosh-lite.com",
						"https://proxy-2-p-mysql.bosh-lite.com"
					],
					"api_username":"admin",
					"api_password":"barfoo",
					"api_force_https": true,
					"skip_ssl_validation": true
				},
				"standalone": {
					"host": "bosh-lite.com",
					"port": 5432,
					"username": "admin",
					"password": "password"
				},
				"test_password": "meowth",
				"timeout_scale": 2.0
			}`

		manifest := `---
jobs:
- name: proxy_z1
  instances: 1
- name: proxy_z2
  instances: 1
- name: proxy_z3
  instances: 1
properties:
  cf:
    api_url: https://api.bosh-lite.com
    app_domains:
    - example.com
    - bosh-lite.com
    admin_username: admin
    admin_password: foobar
    smoke_tests:
      use_existing_org: true
      org: system
    skip_ssl_validation: true
  cf_mysql:
    host: bosh-lite.com
    external_host: p-mysql.bosh-lite.com
    smoke_tests:
      password: meowth
      timeout_scale: 2.0
    broker:
      services:
      - name: p-mysql
        max_user_connections_default: 40
        plans:
        - name: 10mb
          private: true
          max_storage_mb: 10
          max_user_connections: 20
        - name: 100mb
          max_storage_mb: 100
    mysql:
      port: 5432
      admin_username: admin
      admin_password: password
    proxy:
      api_username: admin
      api_password: barfoo
      api_force_https: true`

		err := ioutil.WriteFile(manifestPath, []byte(manifest), 0644)
		Expect(err).NotTo(HaveOccurred())

		configureCmd := exec.Command(
			binPath,
			"-manifestPath", manifestPath,
		)

		var (
			stdOut bytes.Buffer
		)

		sess, err := gexec.Start(configureCmd, &stdOut, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		<-sess.Exited

		Expect(sess.ExitCode()).To(Equal(0))

		Expect(ioutil.ReadAll(&stdOut)).To(MatchJSON(config))

	})
})
