package healthchecker

import (
	"os/exec"
	"regexp"

	"github.com/pivotal-golang/lager"
)

type ScriptHealthChecker struct {
	logger     lager.Logger
	scriptPath string
}

func NewScriptHealthChecker(scriptPath string, logger lager.Logger) *ScriptHealthChecker {
	return &ScriptHealthChecker{
		scriptPath: scriptPath,
		logger:     logger,
	}
}

func (checker *ScriptHealthChecker) Check() bool {
	cmd := exec.Command(checker.scriptPath)
	checker.logger.Info(
		"Executing script",
		lager.Data{"scriptPath": checker.scriptPath},
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		checker.logger.Info(
			"Error executing script",
			lager.Data{"script": checker.scriptPath,
				"error": err.Error(),
			},
		)
		return false
	}

	matchesOne := regexp.MustCompile(`1`)
	return matchesOne.MatchString(string(out[:]))
}
