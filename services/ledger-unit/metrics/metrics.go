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

	localfs "github.com/jancajthaml-openbank/local-fs"
	"github.com/jancajthaml-openbank/ledger-unit/utils"
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics holds metrics counters
type Metrics struct {
	utils.DaemonSupport
	storage                localfs.PlaintextStorage
	tenant                 string
	refreshRate            time.Duration
	promisedTransactions   metrics.Counter
	promisedTransfers      metrics.Counter
	committedTransactions  metrics.Counter
	committedTransfers     metrics.Counter
	rollbackedTransactions metrics.Counter
	rollbackedTransfers    metrics.Counter
	forwardedTransactions  metrics.Counter
	forwardedTransfers     metrics.Counter
}

// NewMetrics returns blank metrics holder
func NewMetrics(ctx context.Context, output string, tenant string, refreshRate time.Duration) Metrics {
	return Metrics{
		DaemonSupport:          utils.NewDaemonSupport(ctx, "metrics"),
		storage:                localfs.NewPlaintextStorage(output),
		tenant:                 tenant,
		refreshRate:            refreshRate,
		promisedTransactions:   metrics.NewCounter(),
		promisedTransfers:      metrics.NewCounter(),
		committedTransactions:  metrics.NewCounter(),
		committedTransfers:     metrics.NewCounter(),
		rollbackedTransactions: metrics.NewCounter(),
		rollbackedTransfers:    metrics.NewCounter(),
		forwardedTransactions:  metrics.NewCounter(),
		forwardedTransfers:     metrics.NewCounter(),
	}
}

// TransactionPromised increments transactions promised by one
func (metrics *Metrics) TransactionPromised(transfers int) {
	metrics.promisedTransactions.Inc(1)
	metrics.promisedTransfers.Inc(int64(transfers))
}

// TransactionCommitted increments transactions committed by one
func (metrics *Metrics) TransactionCommitted(transfers int) {
	metrics.committedTransactions.Inc(1)
	metrics.committedTransfers.Inc(int64(transfers))
}

// TransactionRollbacked increments transactions rollbacked by one
func (metrics *Metrics) TransactionRollbacked(transfers int) {
	metrics.rollbackedTransactions.Inc(1)
	metrics.rollbackedTransfers.Inc(int64(transfers))
}

// TransactionForwarded increments transactions forwarded by one
func (metrics *Metrics) TransactionForwarded(transfers int) {
	metrics.forwardedTransactions.Inc(1)
	metrics.forwardedTransfers.Inc(int64(transfers))
}

// Start handles everything needed to start metrics daemon
func (metrics Metrics) Start() {
	ticker := time.NewTicker(metrics.refreshRate)
	defer ticker.Stop()

	if err := metrics.Hydrate(); err != nil {
		log.Warn(err.Error())
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

	log.Infof("Start metrics daemon, update each %v into %v", metrics.refreshRate, metrics.storage.Root)

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
	log.Info("Stop metrics daemon")
}
