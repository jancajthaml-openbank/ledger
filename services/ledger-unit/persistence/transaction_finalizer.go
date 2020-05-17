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

package persistence

import (
	"context"
	"time"

	"github.com/jancajthaml-openbank/ledger-unit/metrics"
	"github.com/jancajthaml-openbank/ledger-unit/utils"

	system "github.com/jancajthaml-openbank/actor-system"
	localfs "github.com/jancajthaml-openbank/local-fs"
	log "github.com/sirupsen/logrus"
)

// TransactionFinalizer represents journal saturation update subroutine
type TransactionFinalizer struct {
	utils.DaemonSupport
	callback     func(msg interface{}, to system.Coordinates, from system.Coordinates)
	metrics      *metrics.Metrics
	storage      *localfs.PlaintextStorage
	scanInterval time.Duration
}

// NewTransactionFinalizer returns snapshot updater fascade
func NewTransactionFinalizer(ctx context.Context, scanInterval time.Duration, metrics *metrics.Metrics, storage *localfs.PlaintextStorage, callback func(msg interface{}, to system.Coordinates, from system.Coordinates)) TransactionFinalizer {
	return TransactionFinalizer{
		DaemonSupport: utils.NewDaemonSupport(ctx, " transaction-finalizer"),
		callback:      callback,
		metrics:       metrics,
		storage:       storage,
		scanInterval:  scanInterval,
	}
}

func (scan TransactionFinalizer) performIntegrityScan() {
	log.Warn("Transaction finalization not implemented")
}

// Start handles everything needed to start transaction finalizer daemon
func (scan TransactionFinalizer) Start() {
	ticker := time.NewTicker(scan.scanInterval)
	defer ticker.Stop()

	scan.MarkReady()

	select {
	case <-scan.CanStart:
		break
	case <-scan.Done():
		scan.MarkDone()
		return
	}

	log.Infof("Start transaction-finalizer check daemon, scan each %v", scan.scanInterval)

	go func() {
		for {
			select {
			case <-scan.Done():
				scan.MarkDone()
				return
			case <-ticker.C:
				//scan.metrics.TimeFinalizeTransactions(func() {
				scan.performIntegrityScan()
				//})
			}
		}
	}()

	<-scan.IsDone
	log.Info("Stop transaction-finalizer daemon")
}
