package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tedsuo/ifrit"

	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/pivotal-cf-experimental/cf-mysql-quota-enforcer/config"
	"github.com/pivotal-cf-experimental/cf-mysql-quota-enforcer/database"
	"github.com/pivotal-cf-experimental/cf-mysql-quota-enforcer/enforcer"
	"github.com/pivotal-cf-experimental/service-config"
	"github.com/pivotal-golang/lager"
)

type Config struct {
	Host         string
	Port         int
	User         string
	Password     string
	BrokerDBName string
}

func main() {
	serviceConfig := service_config.New()

	flags := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	runOnce := flags.Bool("runOnce", false, "Run only once instead of continuously")
	pidFile := flags.String("pidFile", "", "Location of pid file")
	serviceConfig.AddFlags(flags)
	cf_lager.AddFlags(flags)
	flags.Parse(os.Args[1:])
	logger, _ := cf_lager.New("Quota Enforcer")

	var config config.Config
	err := serviceConfig.Read(&config)
	if err != nil {
		logger.Fatal("Failed to read config", err)
	}

	err = config.Validate()
	if err != nil {
		logger.Fatal("Invalid config", err)
	}

	brokerDBName := config.DBName
	if brokerDBName == "" {
		logger.Fatal("Must specify DBName in the config file", nil)
	}

	db, err := database.NewConnection(config)
	if db != nil {
		defer db.Close()
	}

	if err != nil {
		logger.Fatal("Failed to open database connection", err)
	}

	logger.Info(
		"Database connection established.",
		lager.Data{
			"Host":         config.Host,
			"Port":         config.Port,
			"User":         config.User,
			"DatabaseName": brokerDBName,
		})

	violatorRepo := database.NewViolatorRepo(brokerDBName, db, logger)
	reformerRepo := database.NewReformerRepo(brokerDBName, db, logger)

	e := enforcer.NewEnforcer(violatorRepo, reformerRepo, logger)
	r := enforcer.NewRunner(e, logger)

	if *runOnce {
		logger.Info("Running once")

		err := e.EnforceOnce()
		if err != nil {
			logger.Info(fmt.Sprintf("Quota Enforcing Failed: %s", err.Error()))
		}
	} else {
		process := ifrit.Invoke(r)
		logger.Info("Running continuously")

		// Write pid file once we are running continuously
		if *pidFile != "" {
			pid := os.Getpid()
			err = writePidFile(pid, *pidFile)
			if err != nil {
				logger.Fatal("Cannot write pid to file", err, lager.Data{"pidFile": pidFile, "pid": pid})
			}
			logger.Info("Wrote pid to file", lager.Data{"pidFile": pidFile, "pid": pid})
		}

		err := <-process.Wait()
		if err != nil {
			logger.Fatal("Quota Enforcing Failed", err)
		}
	}
}

func writePidFile(pid int, pidFile string) error {
	return ioutil.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}
