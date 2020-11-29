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
	"time"
)

// ScheduledDaemon represent work happening repeatedly in given interval
type ScheduledDaemon struct {
	Worker
	name     string
	interval time.Duration
}

// NewScheduledDaemon returns new daemon with given name for periodic work
func NewScheduledDaemon(name string, worker Worker, interval time.Duration) Daemon {
	return ScheduledDaemon{
		Worker:   worker,
		name:     name,
		interval: interval,
	}
}

// Done returns signal when worker has finished work
func (daemon ScheduledDaemon) Done() <-chan interface{} {
	return daemon.Worker.Done()
}

// Setup prepares worker for work
func (daemon ScheduledDaemon) Setup() error {
	return daemon.Worker.Setup()
}

// Stop cancels worker's work
func (daemon ScheduledDaemon) Stop() {
	daemon.Worker.Cancel()
}

// Start starts worker's work in given interval and on termination
func (daemon ScheduledDaemon) Start(parentContext context.Context, cancelFunction context.CancelFunc) {
	defer cancelFunction()
	ticker := time.NewTicker(daemon.interval)
	defer ticker.Stop()
	err := daemon.Setup()
	if err != nil {
		log.Error().Msgf("Setup daemon %s error %+v", daemon.name, err.Error())
		return
	}
	log.Info().Msgf("Start daemon %s run each %v", daemon.name, daemon.interval)
	for {
		select {
		case <-parentContext.Done():
			break
		case <-ticker.C:
			daemon.Work()
		}
	}
	daemon.Work()
	daemon.Stop()
	log.Info().Msgf("Stop daemon %s", daemon.name)
}
