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
	localfs "github.com/jancajthaml-openbank/local-fs"
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics holds metrics counters
type Metrics struct {
	storage                         localfs.Storage
	tenant                          string
	continuous                      bool
	promisedTransactions            metrics.Counter
	promisedTransfers               metrics.Counter
	committedTransactions           metrics.Counter
	committedTransfers              metrics.Counter
	rollbackedTransactions          metrics.Counter
	rollbackedTransfers             metrics.Counter
	transactionFinalizerCronLatency metrics.Timer
}

// NewMetrics returns blank metrics holder
func NewMetrics(output string, continuous bool, tenant string) *Metrics {
	storage, err := localfs.NewPlaintextStorage(output)
	if err != nil {
		log.Error().Msgf("Failed to ensure storage %+v", err)
		return nil
	}
	return &Metrics{
		continuous:                      continuous,
		storage:                         storage,
		tenant:                          tenant,
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
		f()
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

// Setup hydrates metrics from storage
func (metrics *Metrics) Setup() error {
	if metrics == nil {
		return nil
	}
	if metrics.continuous {
		metrics.Hydrate()
	}
	return nil
}

// Done returns always finished
func (metrics *Metrics) Done() <-chan interface{} {
	done := make(chan interface{})
	close(done)
	return done
}

// Cancel does nothing
func (metrics *Metrics) Cancel() {
}

// Work represents metrics worker work
func (metrics *Metrics) Work() {
	metrics.Persist()
}
