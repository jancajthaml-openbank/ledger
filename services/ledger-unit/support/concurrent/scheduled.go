// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"time"
)

// ScheduledDaemon represent work happening repeatedly in given interval
type ScheduledDaemon struct {
	Worker
	name       string
	interval   time.Duration
	ticker     *time.Ticker
	cancelOnce sync.Once
	done       chan interface{}
}

// NewScheduledDaemon returns new daemon with given name for periodic work
func NewScheduledDaemon(name string, worker Worker, interval time.Duration) Daemon {
	return &ScheduledDaemon{
		Worker:     worker,
		name:       name,
		interval:   interval,
		ticker:     time.NewTicker(interval),
		cancelOnce: sync.Once{},
		done:       make(chan interface{}),
	}
}

// Done returns signal when worker has finished work
func (daemon *ScheduledDaemon) Done() <-chan interface{} {
	if daemon == nil {
		done := make(chan interface{})
		close(done)
		return done
	}
	<-daemon.Worker.Done()
	return daemon.done
}

// Setup prepares worker for work
func (daemon *ScheduledDaemon) Setup() error {
	if daemon == nil {
		return nil
	}
	return daemon.Worker.Setup()
}

// Stop cancels worker's work
func (daemon *ScheduledDaemon) Stop() {
	if daemon == nil {
		return
	}
	daemon.cancelOnce.Do(func() {
		daemon.ticker.Stop()
		daemon.Worker.Cancel()
		close(daemon.done)
	})
}

// Start starts worker's work in given interval and on termination
func (daemon *ScheduledDaemon) Start(parentContext context.Context, cancelFunction context.CancelFunc) {
	defer cancelFunction()
	if daemon == nil {
		return
	}
	err := daemon.Setup()
	if err != nil {
		log.Error().Err(err).Msgf("Setup daemon %s", daemon.name)
		return
	}

	log.Info().Msgf("Start daemon %s run each %v", daemon.name, daemon.interval)
	defer log.Info().Msgf("Stop daemon %s", daemon.name)

	for {
		select {
		case <-parentContext.Done():
			daemon.Stop()
			return
		case <-daemon.ticker.C:
			daemon.Work()
		case <-daemon.Done():
			return
		}
	}
}
