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

package metrics

import (
	"context"
	"time"

	"github.com/jancajthaml-openbank/ledger-unit/support/concurrent"
	localfs "github.com/jancajthaml-openbank/local-fs"
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics holds metrics counters
type Metrics struct {
	concurrent.DaemonSupport
	storage                         localfs.Storage
	tenant                          string
	refreshRate                     time.Duration
	promisedTransactions            metrics.Counter
	promisedTransfers               metrics.Counter
	committedTransactions           metrics.Counter
	committedTransfers              metrics.Counter
	rollbackedTransactions          metrics.Counter
	rollbackedTransfers             metrics.Counter
	transactionFinalizerCronLatency metrics.Timer
}

// NewMetrics returns blank metrics holder
func NewMetrics(ctx context.Context, output string, tenant string, refreshRate time.Duration) *Metrics {
	storage, err := localfs.NewPlaintextStorage(output)
	if err != nil {
		log.Error().Msgf("Failed to ensure storage %+v", err)
		return nil
	}
	return &Metrics{
		DaemonSupport:                   concurrent.NewDaemonSupport(ctx, "metrics"),
		storage:                         storage,
		tenant:                          tenant,
		refreshRate:                     refreshRate,
		promisedTransactions:            metrics.NewCounter(),
		promisedTransfers:               metrics.NewCounter(),
		committedTransactions:           metrics.NewCounter(),
		committedTransfers:              metrics.NewCounter(),
		rollbackedTransactions:          metrics.NewCounter(),
		rollbackedTransfers:             metrics.NewCounter(),
		transactionFinalizerCronLatency: metrics.NewTimer(),
	}
}

// TimeFinalizeTransactions measures time of finalizeStaleTransactions function run
func (metrics *Metrics) TimeFinalizeTransactions(f func()) {
	if metrics == nil {
		return
	}
	metrics.transactionFinalizerCronLatency.Time(f)
}

// TransactionPromised increments transactions promised by one
func (metrics *Metrics) TransactionPromised(transfers int) {
	if metrics == nil {
		return
	}
	metrics.promisedTransactions.Inc(1)
	metrics.promisedTransfers.Inc(int64(transfers))
}

// TransactionCommitted increments transactions committed by one
func (metrics *Metrics) TransactionCommitted(transfers int) {
	if metrics == nil {
		return
	}
	metrics.committedTransactions.Inc(1)
	metrics.committedTransfers.Inc(int64(transfers))
}

// TransactionRollbacked increments transactions rollbacked by one
func (metrics *Metrics) TransactionRollbacked(transfers int) {
	if metrics == nil {
		return
	}
	metrics.rollbackedTransactions.Inc(1)
	metrics.rollbackedTransfers.Inc(int64(transfers))
}

// Start handles everything needed to start metrics daemon
func (metrics *Metrics) Start() {
	if metrics == nil {
		return
	}
	ticker := time.NewTicker(metrics.refreshRate)
	defer ticker.Stop()

	if err := metrics.Hydrate(); err != nil {
		log.Warn().Msg(err.Error())
	}

	metrics.Persist()
	metrics.MarkReady()

	select {
	case <-metrics.CanStart:
		break
	case <-metrics.Done():
		metrics.MarkDone()
		return
	}

	log.Info().Msgf("Start metrics daemon, update file each %v", metrics.refreshRate)

	go func() {
		for {
			select {
			case <-metrics.Done():
				metrics.Persist()
				metrics.MarkDone()
				return
			case <-ticker.C:
				metrics.Persist()
			}
		}
	}()

	metrics.WaitStop()
	log.Info().Msg("Stop metrics daemon")
}
