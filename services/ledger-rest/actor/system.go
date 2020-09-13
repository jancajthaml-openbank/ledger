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

package actor

import (
	"context"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/metrics"

	system "github.com/jancajthaml-openbank/actor-system"
)

// ActorSystem represents actor system subroutine
type ActorSystem struct {
	system.System
	Metrics *metrics.Metrics
}

// NewActorSystem returns actor system fascade
func NewActorSystem(ctx context.Context, lakeEndpoint string, metrics *metrics.Metrics) ActorSystem {
	result := ActorSystem{
		System:  system.New(ctx, "LedgerRest", lakeEndpoint),
		Metrics: metrics,
	}

	result.System.RegisterOnMessage(ProcessMessage(&result))
	return result
}

// Start daemon noop
func (system ActorSystem) Start() {
	system.System.Start()
}

// Stop daemon noop
func (system ActorSystem) Stop() {
	system.System.Stop()
}

// WaitStop daemon noop
func (system ActorSystem) WaitStop() {
	system.System.WaitStop()
}

// GreenLight daemon noop
func (system ActorSystem) GreenLight() {
	system.System.GreenLight()
}

// WaitReady wait for system to be ready
func (system ActorSystem) WaitReady(deadline time.Duration) error {
	return system.System.WaitReady(deadline)
}
