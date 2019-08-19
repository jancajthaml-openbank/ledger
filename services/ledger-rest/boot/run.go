// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package boot

import (
	"fmt"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/utils"

	log "github.com/sirupsen/logrus"
)

// Stop stops the application
func (prog Program) Stop() {
	close(prog.interrupt)
}

// WaitReady wait for daemons to be ready
func (prog Program) WaitReady(deadline time.Duration) error {
	errors := make([]error, 0)
	mux := new(sync.Mutex)

	var wg sync.WaitGroup
	waitWithDeadline := func(support utils.Daemon) {
		go func() {
			err := support.WaitReady(deadline)
			if err != nil {
				mux.Lock()
				errors = append(errors, err)
				mux.Unlock()
			}
			wg.Done()
		}()
	}

	wg.Add(6)
	waitWithDeadline(prog.actorSystem)
	waitWithDeadline(prog.rest)
	waitWithDeadline(prog.systemControl)
	waitWithDeadline(prog.diskMonitor)
	waitWithDeadline(prog.memoryMonitor)
	waitWithDeadline(prog.metrics)
	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("%+v", errors)
	}

	return nil
}

// GreenLight daemons
func (prog Program) GreenLight() {
	prog.diskMonitor.GreenLight()
	prog.memoryMonitor.GreenLight()
	prog.metrics.GreenLight()
	prog.actorSystem.GreenLight()
	prog.systemControl.GreenLight()
	prog.rest.GreenLight()
}

// WaitInterrupt wait for signal
func (prog Program) WaitInterrupt() {
	<-prog.interrupt
}

// Run runs the application
func (prog Program) Run() {
	go prog.diskMonitor.Start()
	go prog.memoryMonitor.Start()
	go prog.metrics.Start()
	go prog.actorSystem.Start()
	go prog.systemControl.Start()
	go prog.rest.Start()

	if err := prog.WaitReady(5 * time.Second); err != nil {
		log.Errorf("Error when starting daemons: %+v", err)
	} else {
		log.Info(">>> Started <<<")
		utils.NotifyServiceReady()
		prog.GreenLight()
		signal.Notify(prog.interrupt, syscall.SIGINT, syscall.SIGTERM)
		prog.WaitInterrupt()
	}

	log.Info(">>> Stopping <<<")
	utils.NotifyServiceStopping()

	prog.rest.Stop()
	prog.actorSystem.Stop()
	prog.systemControl.Stop()
	prog.diskMonitor.Stop()
	prog.memoryMonitor.Stop()
	prog.metrics.Stop()
	prog.cancel()
}
