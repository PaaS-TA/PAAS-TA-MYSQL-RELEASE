package helpers

import (
	"fmt"
	"os"

	"github.com/cloudfoundry-incubator/cf-test-helpers/services"
)

type Component struct {
	Ip        string `json:"ip"`
	SshTunnel string `json:"ssh_tunnel"`
}

type Plan struct {
	Name               string `json:"plan_name"`
	MaxStorageMb       int    `json:"max_storage_mb"`
	MaxUserConnections int    `json:"max_user_connections"`
}

type Proxy struct {
	ExternalHost      string `json:"external_host"`
	APIUsername       string `json:"api_username"`
	APIPassword       string `json:"api_password"`
	SkipSSLValidation bool   `json:"skip_ssl_validation"`
	ForceHTTPS        bool   `json:"api_force_https"`
}

type MysqlIntegrationConfig struct {
	services.Config
	BrokerHost  string      `json:"broker_host"`
	ServiceName string      `json:"service_name"`
	Plans       []Plan      `json:"plans"`
	Brokers     []Component `json:"brokers"`
	MysqlNodes  []Component `json:"mysql_nodes"`
	Proxy       Proxy       `json:"proxy"`
}

func (c MysqlIntegrationConfig) AppURI(appname string) string {
	return "http://" + appname + "." + c.AppsDomain
}

func LoadConfig() (MysqlIntegrationConfig, error) {
	config := MysqlIntegrationConfig{}

	path := os.Getenv("CONFIG")
	if path == "" {
		return config, fmt.Errorf("Must set $CONFIG to point to an integration config .json file.")
	}

	err := services.LoadConfig(path, &config)
	if err != nil {
		return config, fmt.Errorf("Loading config: %s", err.Error())
	}

	return config, nil
}

func ValidateConfig(config *MysqlIntegrationConfig) error {
	err := services.ValidateConfig(&config.Config)
	if err != nil {
		return err
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

	emptyProxy := Proxy{}
	if config.Proxy == emptyProxy {
		return fmt.Errorf("Field 'proxy' must not be empty")
	}

	if config.Proxy.ExternalHost == "" {
		return fmt.Errorf("Field 'proxy.external_host' must not be empty")
	}

	if config.Proxy.APIUsername == "" {
		return fmt.Errorf("Field 'proxy.api_username' must not be empty")
	}

	if config.Proxy.APIPassword == "" {
		return fmt.Errorf("Field 'proxy.api_password' must not be empty")
	}

	return nil
}
