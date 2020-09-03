package helpers

type MysqlManifest struct {
	Jobs       []Job      `yaml:"jobs"`
	Properties Properties `yaml:"properties"`
}

type Job struct {
	Instances int    `yaml:"instances"`
	Name      string `yaml:"name"`
}

type Properties struct {
	CF struct {
		APIURL        string   `yaml:"api_url"`
		AppDomains    []string `yaml:"app_domains"`
		AdminUsername string   `yaml:"admin_username"`
		AdminPassword string   `yaml:"admin_password"`
		SmokeTests    struct {
			UseExistingOrg bool   `yaml:"use_existing_org"`
			Org            string `yaml:"org"`
		} `yaml:"smoke_tests"`
		SkipSSLValidation bool `yaml:"skip_ssl_validation"`
	} `yaml:"cf"`

	CFMySQL struct {
		Host  string `yaml:"host"`
		MySQL struct {
			Port          int    `yaml:"port"`
			AdminUsername string `yaml:"admin_username"`
			AdminPassword string `yaml:"admin_password"`
		} `yaml:"mysql"`
		SmokeTests struct {
			Password     string  `yaml:"password"`
			TimeoutScale float64 `yaml:"timeout_scale"`
		} `yaml:"smoke_tests"`
		ExternalHost string `yaml:"external_host"`
		Broker       struct {
			Services []struct {
				Name                      string         `yaml:"name"`
				MaxUserConnectionsDefault int            `yaml:"max_user_connections_default"`
				Plans                     []ManifestPlan `yaml:"plans"`
			} `yaml:"services"`
		} `yaml:"broker"`
		Proxy struct {
			APIUsername   string `yaml:"api_username"`
			APIPassword   string `yaml:"api_password"`
			APIForceHTTPS bool   `yaml:"api_force_https"`
		} `yaml:"proxy"`
	} `yaml:"cf_mysql"`
}

type ManifestPlan struct {
	Name               string `yaml:"name"`
	Private            bool   `yaml:"private"`
	MaxStorageMB       int    `yaml:"max_storage_mb"`
	MaxUserConnections int    `yaml:"max_user_connections"`
}
