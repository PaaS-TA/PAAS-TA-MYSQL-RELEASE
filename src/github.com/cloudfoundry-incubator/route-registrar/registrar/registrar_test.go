package registrar_test

import (
	"os"
	"syscall"

	"github.com/cloudfoundry/yagnats"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/route-registrar/config"
	"github.com/cloudfoundry-incubator/route-registrar/healthchecker/fakes"
	. "github.com/cloudfoundry-incubator/route-registrar/registrar"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var config Config
var testSpyClient *yagnats.Client

var _ = Describe("Registrar.RegisterRoutes", func() {
	var logger lager.Logger
	messageBusServer := MessageBusServer{
		"127.0.0.1:4222",
		"nats",
		"nats",
	}

	healthCheckerConfig := HealthCheckerConf{
		Name:     "a_useful_health_checker",
		Interval: 1,
	}

	config = Config{
		[]MessageBusServer{messageBusServer, messageBusServer}, // doesn't matter if these are the same, just want to send a slice
		"riakcs.vcap.me",
		"127.0.0.1",
		8080,
		&healthCheckerConfig,
	}

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("Registrar test")
		testSpyClient = yagnats.NewClient()
		connectionInfo := yagnats.ConnectionInfo{
			messageBusServer.Host,
			messageBusServer.User,
			messageBusServer.Password,
			nil,
		}

		err := testSpyClient.Connect(&connectionInfo)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		testSpyClient.Disconnect()
	})

	It("Sends a router.register message and does not send a router.unregister message", func() {
		// Detect when a router.register message gets sent
		var registered chan (string)
		registered = subscribeToRegisterEvents(func(msg *yagnats.Message) {
			registered <- string(msg.Payload)
		})

		// Detect when an unregister message gets sent
		var unregistered chan (bool)
		unregistered = subscribeToUnregisterEvents(func(msg *yagnats.Message) {
			unregistered <- true
		})

		go func() {
			registrar := NewRegistrar(config, logger)
			registrar.RegisterRoutes()
		}()

		// Assert that we got the right router.register message
		var receivedMessage string
		Eventually(registered, 2).Should(Receive(&receivedMessage))
		Expect(receivedMessage).To(Equal(`{"uris":["riakcs.vcap.me"],"host":"127.0.0.1","port":8080}`))

		// Assert that we never got a router.unregister message
		Consistently(unregistered, 2).ShouldNot(Receive())
	})

	It("Emits a router.unregister message when SIGINT is sent to the registrar's signal channel", func() {
		verifySignalTriggersUnregister(syscall.SIGINT, logger)
	})

	It("Emits a router.unregister message when SIGTERM is sent to the registrar's signal channel", func() {
		verifySignalTriggersUnregister(syscall.SIGTERM, logger)
	})

	Context("When the registrar has a healthchecker", func() {
		It("Emits a router.unregister message when registrar's health check fails, and emits a router.register message when registrar's health check back to normal", func() {
			healthy := fakes.NewFakeHealthChecker()
			healthy.CheckReturns(true)

			unregistered := make(chan string)
			registered := make(chan string)
			var registrar *Registrar

			// Listen for a router.unregister event, then set health status to true, then listen for a router.register event
			subscribeToRegisterEvents(func(msg *yagnats.Message) {
				registered <- string(msg.Payload)

				healthy.CheckReturns(false)

				subscribeToUnregisterEvents(func(msg *yagnats.Message) {
					unregistered <- string(msg.Payload)
				})
			})

			go func() {
				registrar = NewRegistrar(config, logger)
				registrar.AddHealthCheckHandler(healthy)
				registrar.RegisterRoutes()
			}()

			var receivedMessage string
			testTimeout := config.HealthChecker.Interval * 3

			Eventually(registered, testTimeout).Should(Receive(&receivedMessage))
			Expect(receivedMessage).To(Equal(`{"uris":["riakcs.vcap.me"],"host":"127.0.0.1","port":8080}`))

			Eventually(unregistered, testTimeout).Should(Receive(&receivedMessage))
			Expect(receivedMessage).To(Equal(`{"uris":["riakcs.vcap.me"],"host":"127.0.0.1","port":8080}`))
		})
	})
})

func verifySignalTriggersUnregister(signal os.Signal, logger lager.Logger) {
	unregistered := make(chan string)
	returned := make(chan bool)

	var registrar *Registrar

	// Trigger a SIGINT after a successful router.register message
	subscribeToRegisterEvents(func(msg *yagnats.Message) {
		registrar.SignalChannel <- signal
	})

	// Detect when a router.unregister message gets sent
	subscribeToUnregisterEvents(func(msg *yagnats.Message) {
		unregistered <- string(msg.Payload)
	})

	go func() {
		registrar = NewRegistrar(config, logger)
		registrar.RegisterRoutes()

		// Set up a channel to wait for RegisterRoutes to return
		returned <- true
	}()

	// Assert that we got the right router.unregister message as a result of the signal
	var receivedMessage string
	Eventually(unregistered, 2).Should(Receive(&receivedMessage))
	Expect(receivedMessage).To(Equal(`{"uris":["riakcs.vcap.me"],"host":"127.0.0.1","port":8080}`))

	// Assert that RegisterRoutes returned
	Expect(returned).To(Receive())
}

func subscribeToRegisterEvents(callback func(msg *yagnats.Message)) (registerChannel chan string) {
	registerChannel = make(chan string)
	go testSpyClient.Subscribe("router.register", callback)

	return
}

func subscribeToUnregisterEvents(callback func(msg *yagnats.Message)) (unregisterChannel chan bool) {
	unregisterChannel = make(chan bool)
	go testSpyClient.Subscribe("router.unregister", callback)

	return
}
