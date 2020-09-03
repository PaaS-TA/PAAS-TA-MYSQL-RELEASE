package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cloudfoundry-incubator/cf-mysql-acceptance-tests/helpers"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

func main() {
	var (
		manifestPath string
		manifest     helpers.MysqlManifest
	)

	flag.StringVar(&manifestPath, "manifestPath", "", "Path to the manifest yml to parse")
	flag.Parse()

	f, err := os.Open(manifestPath)
	if err != nil {
		panic(err)
	}

	m, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(m, &manifest)
	if err != nil {
		panic(err)
	}

	cfg := makeIntegrationConfig(manifest)

	err = helpers.ValidateConfig(cfg)
	if err != nil {
		panic(err)
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))

	os.Exit(0)
}

func makeIntegrationConfig(manifest helpers.MysqlManifest) *helpers.MysqlIntegrationConfig {
	cfg := &helpers.MysqlIntegrationConfig{}

	p := manifest.Properties

	cfg.ApiEndpoint = p.CF.APIURL
	cfg.AppsDomain = p.CF.AppDomains[0]
	cfg.AdminUser = p.CF.AdminUsername
	cfg.AdminPassword = p.CF.AdminPassword
	cfg.ConfigurableTestPassword = p.CFMySQL.SmokeTests.Password
	cfg.BrokerHost = p.CFMySQL.ExternalHost

	mysqlService := p.CFMySQL.Broker.Services[0]
	cfg.ServiceName = mysqlService.Name

	if p.CF.SmokeTests.UseExistingOrg {
		cfg.OrgName = p.CF.SmokeTests.Org
	}

	for _, plan := range mysqlService.Plans {
		var c int
		if plan.MaxUserConnections == 0 {
			c = mysqlService.MaxUserConnectionsDefault
		} else {
			c = plan.MaxUserConnections
		}

		cfg.Plans = append(cfg.Plans, helpers.Plan{
			Name:               plan.Name,
			Private:            plan.Private,
			MaxStorageMb:       plan.MaxStorageMB,
			MaxUserConnections: c,
		})
	}

	cfg.SkipSSLValidation = p.CF.SkipSSLValidation

	externalHost := p.CFMySQL.ExternalHost
	var counter int
	for _, job := range manifest.Jobs {
		if strings.HasPrefix(job.Name, "proxy") {
			s := fmt.Sprintf("https://proxy-%d-%s", counter, externalHost)
			counter++

			cfg.Proxy.DashboardUrls = append(cfg.Proxy.DashboardUrls, s)
		}
	}

	cfg.Proxy.SkipSSLValidation = p.CF.SkipSSLValidation
	cfg.Proxy.APIUsername = p.CFMySQL.Proxy.APIUsername
	cfg.Proxy.APIPassword = p.CFMySQL.Proxy.APIPassword
	cfg.Proxy.APIForceHTTPS = p.CFMySQL.Proxy.APIForceHTTPS

	cfg.TimeoutScale = p.CFMySQL.SmokeTests.TimeoutScale

	cfg.Standalone.Host = p.CFMySQL.Host
	cfg.Standalone.Port = p.CFMySQL.MySQL.Port
	cfg.Standalone.MySQLUsername = p.CFMySQL.MySQL.AdminUsername
	cfg.Standalone.MySQLPassword = p.CFMySQL.MySQL.AdminPassword

	return cfg
}
