package test_helpers

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"time"
)

func StartNats(port int) *exec.Cmd {
	cmd := exec.Command("gnatsd", "-p", strconv.Itoa(port), "--user", "nats", "--pass", "nats")
	err := cmd.Start()
	if err != nil {
		fmt.Printf("NATS failed to start: %v\n", err)
	}
	err = waitUntilNatsUp(port)
	if err != nil {
		panic("Cannot connect to NATS")
	}
	return cmd
}

func StopCmd(cmd *exec.Cmd) {
	cmd.Process.Kill()
	cmd.Wait()
}

func waitUntilNatsUp(port int) error {
	maxWait := 10
	for i := 0; i < maxWait; i++ {
		time.Sleep(500 * time.Millisecond)
		_, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return nil
		}
	}
	return errors.New("Waited too long for NATS to start")
}
