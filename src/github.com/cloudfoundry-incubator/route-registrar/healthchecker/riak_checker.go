package healthchecker

import (
	"os/exec"
	"regexp"

	"github.com/pivotal-golang/lager"
)

type RiakHealthChecker struct {
	logger           lager.Logger
	status           bool
	pidFileName      string
	riakAdminProgram string
	nodeIpAddress    string
	scriptPath       string
}

func (checker *RiakHealthChecker) Check() bool {
	pidFileExists := checkPIDExist(checker.pidFileName)
	checker.status = pidFileExists

	if pidFileExists {
		nodeExistsAndIsValid := checker.nodeExistsAndIsValid(checker.nodeIpAddress)
		checker.status = checker.status && nodeExistsAndIsValid

		if !nodeExistsAndIsValid {
			checker.logger.Info(
				"Riak node is not a valid member of the cluster",
				lager.Data{"nodeIpAddress": checker.nodeIpAddress},
			)
		}
	} else {
		checker.logger.Info(
			"Riak pidFile does not exist",
			lager.Data{"pidFile": checker.pidFileName},
		)
	}

	return checker.status
}

func NewRiakHealthChecker(pidFileName string, riakAdminProgram string, nodeIpAddress string, logger lager.Logger, scriptPath string) *RiakHealthChecker {
	return &RiakHealthChecker{
		status:           false,
		pidFileName:      pidFileName,
		riakAdminProgram: riakAdminProgram,
		nodeIpAddress:    nodeIpAddress,
		logger:           logger,
		scriptPath:       scriptPath,
	}
}

func (checker *RiakHealthChecker) nodeExistsAndIsValid(nodeIp string) (result bool) {
	cmd := exec.Command(checker.scriptPath, checker.riakAdminProgram, nodeIp)

	out, err := cmd.CombinedOutput()
	if err != nil {
		checker.logger.Info(
			"Error checking node validity",
			lager.Data{
				"nodeValidityCheckerProgram": checker.scriptPath,
				"err": err,
			},
		)
		return false
	}

	matchesOne := regexp.MustCompile(`1`)
	return matchesOne.MatchString(string(out[:]))
}
