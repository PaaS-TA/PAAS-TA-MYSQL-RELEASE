package main

import (
	"flag"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/route-registrar/config"
	. "github.com/cloudfoundry-incubator/route-registrar/healthchecker"
	. "github.com/cloudfoundry-incubator/route-registrar/registrar"
	"github.com/pivotal-cf-experimental/service-config"
	"github.com/pivotal-golang/lager"
)

func main() {
	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	pidfile := flags.String("pidfile", "", "Path to pid file")
	scriptPath := flags.String("scriptPath", "./check_node_validity.sh", "Path to script file")
	cf_lager.AddFlags(flags)

	serviceConfig := service_config.New()
	serviceConfig.AddFlags(flags)
	flags.Set("configPath", "registrar_settings.yml")

	flags.Parse(os.Args[1:])

	logger, _ := cf_lager.New("Route Registrar")

	logger.Info("Route Registrar started")

	var registrarConfig config.Config
	err := serviceConfig.Read(&registrarConfig)
	if err != nil {
		logger.Fatal("error parsing file: %s\n", err)
	}

	registrar := NewRegistrar(registrarConfig, logger)
	//add health check handler
	checker := InitHealthChecker(registrarConfig, logger, *scriptPath)
	if checker != nil {
		registrar.AddHealthCheckHandler(checker)
	}

	if *pidfile != "" {
		pid := strconv.Itoa(os.Getpid())
		err := ioutil.WriteFile(*pidfile, []byte(pid), 0644)
		if err != nil {
			logger.Fatal(
				"error writing pid to pidfile",
				err,
				lager.Data{
					"pid":     pid,
					"pidfile": *pidfile},
			)
		}
	}

	registrar.RegisterRoutes()
	os.Exit(1)
}
