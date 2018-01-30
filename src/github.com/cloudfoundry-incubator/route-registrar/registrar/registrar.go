package registrar

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudfoundry/gibson"
	"github.com/cloudfoundry/yagnats"

	"github.com/cloudfoundry-incubator/route-registrar/config"
	. "github.com/cloudfoundry-incubator/route-registrar/healthchecker"

	"github.com/pivotal-golang/lager"
)

type Registrar struct {
	logger               lager.Logger
	Config               config.Config
	SignalChannel        chan os.Signal
	HealthChecker        HealthChecker
	previousHealthStatus bool
}

func NewRegistrar(clientConfig config.Config, logger lager.Logger) *Registrar {
	return &Registrar{
		Config:               clientConfig,
		logger:               logger,
		SignalChannel:        make(chan os.Signal, 1),
		previousHealthStatus: false,
	}
}

func (registrar *Registrar) AddHealthCheckHandler(handler HealthChecker) {
	registrar.HealthChecker = handler
}

type callbackFunction func()

func (registrar *Registrar) RegisterRoutes() {
	messageBus := buildMessageBus(registrar)
	client := gibson.NewCFRouterClient(registrar.Config.ExternalIp, messageBus)

	// set up periodic registration
	client.Greet()

	done := make(chan bool)
	registrar.registerSignalHandler(done, client)

	if registrar.HealthChecker != nil {
		callbackInterval := time.Duration(registrar.Config.HealthChecker.Interval) * time.Second
		callbackPeriodically(callbackInterval,
			func() { registrar.updateRegistrationBasedOnHealthCheck(client) },
			done)
	} else {
		client.Register(registrar.Config.Port, registrar.Config.ExternalHost)

		select {
		case <-done:
			return
		}
	}
}

func buildMessageBus(registrar *Registrar) (messageBus yagnats.NATSClient) {

	messageBus = yagnats.NewClient()
	natsServers := []yagnats.ConnectionProvider{}

	for _, server := range registrar.Config.MessageBusServers {
		registrar.logger.Info(
			"Adding NATS server",
			lager.Data{"server": server},
		)
		natsServers = append(natsServers, &yagnats.ConnectionInfo{
			server.Host,
			server.User,
			server.Password,
			nil,
		})
	}

	natsInfo := &yagnats.ConnectionCluster{natsServers}

	err := messageBus.Connect(natsInfo)

	if err != nil {
		registrar.logger.Info(
			"Error connecting to NATS",
			lager.Data{"error": err.Error()},
		)
		panic("Failed to connect to NATS bus.")
	}

	registrar.logger.Info("Successfully connected to NATS.")

	return
}

func callbackPeriodically(duration time.Duration, callback callbackFunction, done chan bool) {
	interval := time.NewTicker(duration)
	for stop := false; !stop; {
		select {
		case <-interval.C:
			callback()
		case stop = <-done:
			return
		}
	}
}

func (registrar *Registrar) updateRegistrationBasedOnHealthCheck(client *gibson.CFRouterClient) {
	current := registrar.HealthChecker.Check()
	if (!current) && registrar.previousHealthStatus {
		registrar.logger.Info("Health check status changed to unavailabile; unregistering the route")
		client.Unregister(registrar.Config.Port, registrar.Config.ExternalHost)
	} else if current && (!registrar.previousHealthStatus) {
		registrar.logger.Info("Health check status changed to availabile; registering the route")
		client.Register(registrar.Config.Port, registrar.Config.ExternalHost)
	}
	registrar.previousHealthStatus = current
}

func (registrar *Registrar) registerSignalHandler(done chan bool, client *gibson.CFRouterClient) {
	go func() {
		select {
		case <-registrar.SignalChannel:
			registrar.logger.Info("Received SIGTERM or SIGINT; unregistering the route")
			client.Unregister(registrar.Config.Port, registrar.Config.ExternalHost)
			done <- true
		}
	}()

	signal.Notify(registrar.SignalChannel, syscall.SIGINT, syscall.SIGTERM)
}
