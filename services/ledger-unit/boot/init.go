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
	"os"
	"github.com/rs/xid"

	"github.com/jancajthaml-openbank/ledger-unit/actor"
	"github.com/jancajthaml-openbank/ledger-unit/config"
	"github.com/jancajthaml-openbank/ledger-unit/metrics"
	"github.com/jancajthaml-openbank/ledger-unit/model"
	"github.com/jancajthaml-openbank/ledger-unit/support/concurrent"
	"github.com/jancajthaml-openbank/ledger-unit/support/logging"

	system "github.com/jancajthaml-openbank/actor-system"
)

// Program encapsulate program
type Program struct {
	interrupt chan os.Signal
	cfg       config.Configuration
	daemons   []concurrent.Daemon
}

// Register daemon into program
func (prog *Program) Register(daemon concurrent.Daemon) {
	if prog == nil || daemon == nil {
		return
	}
	prog.daemons = append(prog.daemons, daemon)
}

// NewProgram returns new program
func NewProgram() Program {
	return Program{
		interrupt: make(chan os.Signal, 1),
		cfg:       config.LoadConfig(),
		daemons:   make([]concurrent.Daemon, 0),
	}
}

// Setup setups program
func (prog *Program) Setup() {
	if prog == nil {
		return
	}

	logging.SetupLogger(prog.cfg.LogLevel)

	metricsWorker := metrics.NewMetrics(
		prog.cfg.MetricsOutput,
		prog.cfg.MetricsContinuous,
		prog.cfg.Tenant,
	)

	actorSystem := actor.NewActorSystem(
		prog.cfg.Tenant,
		prog.cfg.LakeHostname,
		prog.cfg.RootStorage,
		metricsWorker,
	)

	transactionFinalizerWorker := actor.NewTransactionFinalizer(
		prog.cfg.RootStorage,
		metricsWorker,
		func(transaction model.Transaction) {
			name := "transaction/" + xid.New().String()
			ref, err := actor.NewTransactionActor(actorSystem, name)
			if err != nil {
				return
			}
			ref.Tell(
				transaction,
				system.Coordinates{
					Region: actorSystem.Name,
					Name:   name,
				},
				system.Coordinates{
					Region: actorSystem.Name,
					Name:   "transaction_finalizer_cron",
				},
			)
		},
	)

	prog.Register(concurrent.NewOneShotDaemon(
		"actor-system",
		actorSystem,
	))

	prog.Register(concurrent.NewScheduledDaemon(
		"metrics",
		metricsWorker,
		prog.cfg.MetricsRefreshRate,
	))

	prog.Register(concurrent.NewScheduledDaemon(
		"transaction-finalizer",
		transactionFinalizerWorker,
		prog.cfg.TransactionIntegrityScanInterval,
	))
}
