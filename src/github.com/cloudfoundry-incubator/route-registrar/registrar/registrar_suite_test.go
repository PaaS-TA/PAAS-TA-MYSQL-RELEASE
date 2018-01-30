package registrar_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/route-registrar/test_helpers"
)

func TestRoute_register(t *testing.T) {
	RegisterFailHandler(Fail)

	fmt.Println("starting gnatsd...")
	natsCmd := StartNats(4222)

	RunSpecs(t, "Registrar Suite")

	StopCmd(natsCmd)
}
