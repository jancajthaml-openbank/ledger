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
	"time"

	log "github.com/sirupsen/logrus"
)

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
	metrics.MarkReady()

	select {
	case <-metrics.CanStart:
		break
	case <-metrics.Done():
		metrics.MarkDone()
		return
	}

	log.Infof("Start metrics daemon, update each %v into %v", metrics.refreshRate, metrics.output)

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

	<-metrics.IsDone
	log.Info("Stop metrics daemon")
}
