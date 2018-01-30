package enforcer

import (
	"fmt"
	"os"
	"time"

	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
)

type runner struct {
	enforcer Enforcer
	logger   lager.Logger
}

func NewRunner(enforcer Enforcer, logger lager.Logger) ifrit.Runner {
	return &runner{
		enforcer: enforcer,
		logger:   logger,
	}
}

func (r runner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	go func() {
		for {
			err := r.enforcer.EnforceOnce()
			if err != nil {
				r.logger.Info(fmt.Sprintf("Enforcing Failed: %s", err.Error()))
			}
			time.Sleep(1 * time.Second)
		}
	}()

	close(ready)
	<-signals
	return nil
}
