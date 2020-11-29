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

import "context"

// OneShotDaemon represent work happening only once
type OneShotDaemon struct {
	Worker
	name string
}

// NewOneShotDaemon returns new daemon with given name for single work
func NewOneShotDaemon(name string, worker Worker) Daemon {
	return OneShotDaemon{
		Worker: worker,
		name:   name,
	}
}

// Done returns signal when worker has finished work
func (daemon OneShotDaemon) Done() <-chan interface{} {
	return daemon.Worker.Done()
}

// Setup prepares worker for work
func (daemon OneShotDaemon) Setup() error {
	return daemon.Worker.Setup()
}

// Stop cancels worker's work
func (daemon OneShotDaemon) Stop() {
	daemon.Worker.Cancel()
}

// Start starts worker's work once
func (daemon OneShotDaemon) Start(parentContext context.Context, cancelFunction context.CancelFunc) {
	defer cancelFunction()
	err := daemon.Setup()
	if err != nil {
		log.Error().Msgf("Setup daemon %s error %+v", daemon.name, err.Error())
		return
	}
	go func() {
		<-parentContext.Done()
		daemon.Stop()
	}()
	log.Info().Msgf("Start daemon %s run once", daemon.name)
	daemon.Work()
	<-daemon.Done()
	log.Info().Msgf("Stop daemon %s", daemon.name)
}
