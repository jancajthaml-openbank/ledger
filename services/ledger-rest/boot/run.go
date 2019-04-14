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

	"github.com/jancajthaml-openbank/ledger-rest/daemon"
	"github.com/jancajthaml-openbank/ledger-rest/utils"

	log "github.com/sirupsen/logrus"
)

// Stop stops the application
func (app Application) Stop() {
	close(app.interrupt)
}

// WaitReady wait for daemons to be ready
func (app Application) WaitReady(deadline time.Duration) error {
	errors := make([]error, 0)
	mux := new(sync.Mutex)

	var wg sync.WaitGroup
	waitWithDeadline := func(support daemon.Daemon) {
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

	wg.Add(4)
	waitWithDeadline(app.actorSystem)
	waitWithDeadline(app.rest)
	waitWithDeadline(app.systemControl)
	waitWithDeadline(app.metrics)
	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("%+v", errors)
	}

	return nil
}

// GreenLight daemons
func (app Application) GreenLight() {
	app.metrics.GreenLight()
	app.actorSystem.GreenLight()
	app.systemControl.GreenLight()
	app.rest.GreenLight()
}

// WaitInterrupt wait for signal
func (app Application) WaitInterrupt() {
	<-app.interrupt
}

// Run runs the application
func (app Application) Run() {
	log.Info(">>> Start <<<")

	go app.metrics.Start()
	go app.actorSystem.Start()
	go app.systemControl.Start()
	go app.rest.Start()

	if err := app.WaitReady(5 * time.Second); err != nil {
		log.Errorf("Error when starting daemons: %+v", err)
	} else {
		log.Info(">>> Started <<<")
		utils.NotifyServiceReady()
		app.GreenLight()
		signal.Notify(app.interrupt, syscall.SIGINT, syscall.SIGTERM)
		app.WaitInterrupt()
	}

	log.Info(">>> Stopping <<<")
	utils.NotifyServiceStopping()

	app.rest.Stop()
	app.actorSystem.Stop()
	app.systemControl.Stop()
	app.metrics.Stop()
	app.cancel()

	log.Info(">>> Stop <<<")
}
