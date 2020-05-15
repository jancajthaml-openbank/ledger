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

package actor

import (
	"context"
	"fmt"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/metrics"

	system "github.com/jancajthaml-openbank/actor-system"
)

// ActorSystem represents actor system subroutine
type ActorSystem struct {
	system.Support
	Metrics *metrics.Metrics
}

// NewActorSystem returns actor system fascade
func NewActorSystem(ctx context.Context, lakeEndpoint string, metrics *metrics.Metrics) ActorSystem {
	result := ActorSystem{
		Support: system.NewSupport(ctx, "LedgerRest", lakeEndpoint),
		Metrics: metrics,
	}

	result.Support.RegisterOnRemoteMessage(ProcessRemoteMessage(&result))

	return result
}

// Start daemon noop
func (system ActorSystem) Start() {
	system.Support.Start()
}

// Stop daemon noop
func (system ActorSystem) Stop() {
	system.Support.Stop()
}

// WaitReady wait for system to be ready
func (system ActorSystem) WaitReady(deadline time.Duration) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("unknown panic")
			}
		}
	}()

	ticker := time.NewTicker(deadline)
	select {
	case <-system.Support.IsReady:
		ticker.Stop()
		err = nil
		return
	case <-ticker.C:
		err = fmt.Errorf("actor-system was not ready within %v seconds", deadline)
		return
	}
}
