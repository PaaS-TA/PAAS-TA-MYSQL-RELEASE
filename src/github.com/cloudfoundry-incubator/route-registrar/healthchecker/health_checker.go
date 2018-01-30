package healthchecker

import (
	"os"

	. "github.com/cloudfoundry-incubator/route-registrar/config"
	"github.com/pivotal-golang/lager"
)

type HealthChecker interface {
	Check() bool
}

func InitHealthChecker(clientConfig Config, logger lager.Logger, scriptPath string) HealthChecker {
	if clientConfig.HealthChecker != nil {
		if clientConfig.HealthChecker.Name == "riak-cs-cluster" {
			//TODO: these should be passed in via registrar_settings.yml
			riak_pidfile := "/var/vcap/sys/run/riak/riak.pid"
			riak_admin_program := "/var/vcap/packages/riak/rel/bin/riak-admin"
			riak_cs_pidfile := "/var/vcap/sys/run/riak-cs/riak-cs.pid"

			riakChecker := NewRiakHealthChecker(riak_pidfile, riak_admin_program, clientConfig.ExternalIp, logger, scriptPath)
			riakCSChecker := NewRiakCSHealthChecker(riak_cs_pidfile, logger)
			checkers := []HealthChecker{riakChecker, riakCSChecker}

			checker := NewCompositeChecker(checkers)
			return checker
		}

		if clientConfig.HealthChecker.Name == "script" {
			return NewScriptHealthChecker(clientConfig.HealthChecker.HealthcheckScript, logger)
		}
	}
	return nil
}

func checkPIDExist(pidFileName string) bool {
	_, err := os.Stat(pidFileName)
	return nil == err
}
