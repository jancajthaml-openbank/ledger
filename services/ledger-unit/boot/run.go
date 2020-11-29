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
	"context"
	"github.com/jancajthaml-openbank/ledger-unit/support/concurrent"
	"github.com/jancajthaml-openbank/ledger-unit/support/host"
	"os/signal"
	"sync"
	"syscall"
)

// Done retutrns signal when all daemons are done
func (prog Program) Done() <-chan interface{} {
	out := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(len(prog.daemons))
	for idx := range prog.daemons {
		go func(c concurrent.Daemon) {
			for v := range c.Done() {
				out <- v
			}
			wg.Done()
		}(prog.daemons[idx])
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Stop stops all daemons
func (prog Program) Stop() {
	for idx := range prog.daemons {
		prog.daemons[idx].Stop()
	}
	close(prog.interrupt)
}

// Start starts all daemons and blocks until INT or TERM signal is received
func (prog Program) Start(parentContext context.Context, cancelFunction context.CancelFunc) {
	for idx := range prog.daemons {
		go prog.daemons[idx].Start(parentContext, cancelFunction)
	}
	host.NotifyServiceReady()
	log.Info().Msg("Program Started")
	signal.Notify(prog.interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-prog.interrupt
	log.Info().Msg("Program Stopping")
	if err := host.NotifyServiceStopping(); err != nil {
		log.Error().Msg(err.Error())
	}
	cancelFunction()
	<-prog.Done()
}
