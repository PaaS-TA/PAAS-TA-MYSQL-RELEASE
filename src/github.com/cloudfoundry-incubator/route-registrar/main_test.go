package main_test

import (
	"fmt"
	"os/exec"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {

	var natsCmd *exec.Cmd

	BeforeEach(func() {

		initConfig()
		writeConfig()

		natsCmd = exec.Command(
			"gnatsd",
			"-p", strconv.Itoa(natsPort),
			"--user", "nats",
			"--pass", "nats")
		err := natsCmd.Start()
		Ω(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		natsCmd.Process.Kill()
		natsCmd.Wait()
	})

	It("Starts correctly and exits 1 on SIGTERM", func() {
		command := exec.Command(
			routeRegistrarBinPath,
			fmt.Sprintf("-pidfile=%s", pidFile),
			fmt.Sprintf("-configPath=%s", configFile),
			fmt.Sprintf("-scriptPath=%s", scriptPath),
		)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())

		Eventually(session.Out).Should(gbytes.Say("Route Registrar"))

		time.Sleep(500 * time.Millisecond)

		session.Terminate().Wait()
		Eventually(session).Should(gexec.Exit(1))
	})
})
