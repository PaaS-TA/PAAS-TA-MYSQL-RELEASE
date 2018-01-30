package config

type MessageBusServer struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type HealthCheckerConf struct {
	Name              string  `yaml:"name"`
	Interval          float64 `yaml:"interval_in_seconds"`
	HealthcheckScript string  `yaml:"healthcheck_script_path"`
}

type Config struct {
	MessageBusServers []MessageBusServer `yaml:"message_bus_servers"`
	ExternalHost      string             `yaml:"external_host"`
	ExternalIp        string             `yaml:"external_ip"`
	Port              int                `yaml:"port"`
	HealthChecker     *HealthCheckerConf `yaml:"health_checker"`
}
