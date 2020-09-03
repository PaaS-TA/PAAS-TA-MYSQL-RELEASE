package helpers

import (
	"fmt"
	"os"

	"encoding/json"
	"io/ioutil"

	"github.com/cloudfoundry-incubator/cf-test-helpers/config"
)

type Component struct {
	Ip        string `json:"ip"`
	SshTunnel string `json:"ssh_tunnel"`
}

type Plan struct {
	Name               string `json:"name"`
	MaxStorageMb       int    `json:"max_storage_mb"`
	MaxUserConnections int    `json:"max_user_connections"`
	Private            bool   `json:"private,omitempty"`
}

type Proxy struct {
	DashboardUrls     []string `json:"dashboard_urls"`
	APIUsername       string   `json:"api_username"`
	APIPassword       string   `json:"api_password"`
	SkipSSLValidation bool     `json:"skip_ssl_validation"`
	APIForceHTTPS     bool     `json:"api_force_https"`
}

type Standalone struct {
	Host          string `json:"host"`
	MySQLUsername string `json:"username"`
	MySQLPassword string `json:"password"`
	Port          int    `json:"port"`
}

type Tuning struct {
	ExpectationFilePath string `json:"expectation_file_path"`
}

type MysqlIntegrationConfig struct {
	CFConfig       *config.Config
	BOSH           BOSH        `json:"bosh"`
	BrokerHost     string      `json:"broker_host,omitempty"`
	BrokerProtocol string      `json:"broker_protocol,omitempty"`
	ServiceName    string      `json:"service_name"`
	EnableTlsTests bool        `json:"enable_tls_tests"`
	Plans          []Plan      `json:"plans"`
	Brokers        []Component `json:"brokers,omitempty"`
	MysqlNodes     []Component `json:"mysql_nodes,omitempty"`
	Proxy          Proxy       `json:"proxy"`
	Standalone     Standalone  `json:"standalone,omitempty"`
	StandaloneOnly bool        `json:"standalone_only,omitempty"`
	Tuning         Tuning      `json:"tuning,omitempty"`
}

type BOSH struct {
	CACert       string `json:"ca_cert"`
	Client       string `json:"client"`
	ClientSecret string `json:"client_secret"`
	URL          string `json:"url"`
}

func (c MysqlIntegrationConfig) AppURI(appname string) string {
	return "https://" + appname + "." + c.CFConfig.AppsDomain
}

func LoadConfig() (MysqlIntegrationConfig, error) {
	mysqlIntegrationConfig := MysqlIntegrationConfig{}

	path := os.Getenv("CONFIG")
	if path == "" {
		return mysqlIntegrationConfig, fmt.Errorf("Must set $CONFIG to point to an integration config .json file.")
	}

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(buf, &mysqlIntegrationConfig); err != nil {
		panic(err)
	}

	cfConfig := &config.Config{
		NamePrefix: "MySQLATS",
	}
	if !mysqlIntegrationConfig.StandaloneOnly {
		err = config.Load(path, cfConfig)
		if err != nil {
			return mysqlIntegrationConfig, fmt.Errorf("Loading config: %s", err.Error())
		}
		mysqlIntegrationConfig.CFConfig = cfConfig
	}


	if mysqlIntegrationConfig.BrokerProtocol == "" {
		mysqlIntegrationConfig.BrokerProtocol = "https"
	}

	return mysqlIntegrationConfig, nil
}

func ValidateConfig(config *MysqlIntegrationConfig) error {
	if config.StandaloneOnly {
		if config.Standalone.Host == "" {
			return fmt.Errorf("Field 'standalone.host' must not be empty")
		}

		if config.Standalone.Port == 0 {
			return fmt.Errorf("Field 'standalone.port' must not be empty")
		}

		if config.Standalone.MySQLUsername == "" {
			return fmt.Errorf("Field 'standalone.username' must not be empty")
		}

		if config.Standalone.MySQLPassword == "" {
			return fmt.Errorf("Field 'standalone.password' must not be empty")
		}

		return nil
	}

	if config.ServiceName == "" {
		return fmt.Errorf("Field 'service_name' must not be empty")
	}

	if config.Plans == nil {
		return fmt.Errorf("Field 'plans' must not be nil")
	}

	if len(config.Plans) == 0 {
		return fmt.Errorf("Field 'plans' must not be empty")
	}

	for index, plan := range config.Plans {
		if plan.Name == "" {
			return fmt.Errorf("Field 'plans[%d].name' must not be empty", index)
		}

		if plan.MaxStorageMb == 0 {
			return fmt.Errorf("Field 'plans[%d].max_storage_mb' must not be empty", index)
		}

		if plan.MaxUserConnections == 0 {
			return fmt.Errorf("Field 'plans[%d].max_user_connections' must not be empty", index)
		}
	}

	if config.BrokerHost == "" {
		return fmt.Errorf("Field 'broker_host' must not be empty")
	}

	if len(config.Proxy.DashboardUrls) == 0 {
		return fmt.Errorf("Field 'proxy.dashboardUrls' must not be empty")
	}

	for index, url := range config.Proxy.DashboardUrls {
		if url == "" {
			return fmt.Errorf("Field 'proxy.dashboard_urls[%d]' must not be empty", index)
		}
	}

	if config.Proxy.APIUsername == "" {
		return fmt.Errorf("Field 'proxy.api_username' must not be empty")
	}

	if config.Proxy.APIPassword == "" {
		return fmt.Errorf("Field 'proxy.api_password' must not be empty")
	}

	return nil
}
