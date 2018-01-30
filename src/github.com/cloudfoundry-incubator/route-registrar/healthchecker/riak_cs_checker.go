package healthchecker

import "github.com/pivotal-golang/lager"

type RiakCSHealthChecker struct {
	status      bool
	pidFileName string
	logger      lager.Logger
}

func (checker *RiakCSHealthChecker) Check() bool {
	pidFileExists := checkPIDExist(checker.pidFileName)

	if !pidFileExists {
		checker.logger.Info(
			"RiakCS pidFile does not exist",
			lager.Data{"pidFile": checker.pidFileName},
		)
	}

	checker.status = pidFileExists
	return checker.status
}

func NewRiakCSHealthChecker(pidFileName string, logger lager.Logger) *RiakCSHealthChecker {
	return &RiakCSHealthChecker{
		status:      false,
		pidFileName: pidFileName,
		logger:      logger,
	}
}
