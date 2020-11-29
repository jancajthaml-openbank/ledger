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

package concurrent

import (
	"context"
	"sync"
)

// DaemonPool represent 1:N daemon mapping
type DaemonPool struct {
	name    string
	daemons []Daemon
}

// NewDaemonPool returns new pool
func NewDaemonPool(name string) DaemonPool {
	return DaemonPool{
		name:    name,
		daemons: make([]Daemon, 0),
	}
}

// Register daemon into program
func (pool *DaemonPool) Register(daemon Daemon) {
	if pool == nil || daemon == nil {
		return
	}
	pool.daemons = append(pool.daemons, daemon)
}

// Done aggregates all done signals of daemons into one
func (pool DaemonPool) Done() <-chan interface{} {
	out := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(len(pool.daemons))
	for idx := range pool.daemons {
		go func(daemon Daemon) {
			<-daemon.Done()
			wg.Done()
		}(pool.daemons[idx])
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Stop stops all daemons inside pool
func (pool DaemonPool) Stop() {
	var wg sync.WaitGroup
	wg.Add(len(pool.daemons))
	for idx := range pool.daemons {
		go func(daemon Daemon) {
			daemon.Stop()
			wg.Done()
		}(pool.daemons[idx])
	}
	wg.Wait()
}

// Start starts all daemon in order they were registered
func (pool DaemonPool) Start(parentContext context.Context, cancelFunction context.CancelFunc) {
	defer cancelFunction()
	go func() {
		<-parentContext.Done()
		pool.Stop()
	}()

	log.Info().Msgf("Start pool %s", pool.name)
	var wg sync.WaitGroup
	wg.Add(len(pool.daemons))
	for idx := range pool.daemons {
		go func(daemon Daemon) {
			daemon.Start(parentContext, cancelFunction)
			wg.Done()
		}(pool.daemons[idx])
	}
	wg.Wait()
	log.Info().Msgf("Stop pool %s", pool.name)
}
