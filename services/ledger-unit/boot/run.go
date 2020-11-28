// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
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

	"github.com/jancajthaml-openbank/ledger-unit/support/concurrent"
	"github.com/jancajthaml-openbank/ledger-unit/support/host"
)

// WaitReady wait for daemons to be ready
func (prog Program) WaitReady(deadline time.Duration) error {
	errors := make([]error, 0)
	mux := new(sync.Mutex)

	var wg sync.WaitGroup
	waitWithDeadline := func(support concurrent.Daemon) {
		if support == nil {
			wg.Done()
			return
		}
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

	wg.Add(len(prog.daemons))
	for idx := range prog.daemons {
		waitWithDeadline(prog.daemons[idx])
	}
	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("%+v", errors)
	}

	return nil
}

// GreenLight daemons
func (prog Program) GreenLight() {
	for idx := range prog.daemons {
		if prog.daemons[idx] == nil {
			continue
		}
		prog.daemons[idx].GreenLight()
	}
}

// WaitStop wait for daemons to stop
func (prog Program) WaitStop() {
	for idx := range prog.daemons {
		if prog.daemons[idx] == nil {
			continue
		}
		prog.daemons[idx].WaitStop()
	}
}

// WaitInterrupt wait for signal
func (prog Program) WaitInterrupt() {
	<-prog.interrupt
}

// Stop stops the application
func (prog Program) Stop() {
	close(prog.interrupt)
}

// Start runs the application
func (prog Program) Start() {
	for idx := range prog.daemons {
		if prog.daemons[idx] == nil {
			continue
		}
		go prog.daemons[idx].Start()
	}

	if err := prog.WaitReady(5 * time.Second); err != nil {
		log.Error().Msgf("Error when starting daemons: %+v", err)
	} else {
		host.NotifyServiceReady()
		prog.GreenLight()
		log.Info().Msg("Program Started")
		signal.Notify(prog.interrupt, syscall.SIGINT, syscall.SIGTERM)
		prog.WaitInterrupt()
	}

	log.Info().Msg("Program Stopping")
	if err := host.NotifyServiceStopping(); err != nil {
		log.Error().Msg(err.Error())
	}

	prog.cancel()
	prog.WaitStop()
}
