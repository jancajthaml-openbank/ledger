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

package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/utils"

	metrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

// Metrics represents metrics subroutine
type Metrics struct {
	utils.DaemonSupport
	output                   string
	refreshRate              time.Duration
	getTransactionLatency    metrics.Timer
	getTransactionsLatency   metrics.Timer
	createTransactionLatency metrics.Timer
	forwardTransferLatency   metrics.Timer
}

// NewMetrics returns metrics fascade
func NewMetrics(ctx context.Context, output string, refreshRate time.Duration) Metrics {
	return Metrics{
		DaemonSupport:            utils.NewDaemonSupport(ctx),
		output:                   output,
		refreshRate:              refreshRate,
		getTransactionLatency:    metrics.NewTimer(),
		getTransactionsLatency:   metrics.NewTimer(),
		createTransactionLatency: metrics.NewTimer(),
		forwardTransferLatency:   metrics.NewTimer(),
	}
}

// TimeForwardTransfer measure execution of ForwardTransfer
func (metrics *Metrics) TimeForwardTransfer(f func()) {
	metrics.forwardTransferLatency.Time(f)
}

// TimeGetTransaction measure execution of GetTransaction
func (metrics *Metrics) TimeGetTransaction(f func()) {
	metrics.getTransactionLatency.Time(f)
}

// TimeGetTransactions measure execution of GetTransaction
func (metrics *Metrics) TimeGetTransactions(f func()) {
	metrics.getTransactionsLatency.Time(f)
}

// TimeCreateTransaction measure execution of CreateTransaction
func (metrics *Metrics) TimeCreateTransaction(f func()) {
	metrics.createTransactionLatency.Time(f)
}

// WaitReady wait for metrics to be ready
func (metrics Metrics) WaitReady(deadline time.Duration) (err error) {
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
	case <-metrics.IsReady:
		ticker.Stop()
		err = nil
		return
	case <-ticker.C:
		err = fmt.Errorf("daemon was not ready within %v seconds", deadline)
		return
	}
}

// Start handles everything needed to start metrics daemon
func (metrics Metrics) Start() {
	defer metrics.MarkDone()

	if metrics.output == "" {
		log.Warnf("no metrics output defined, skipping metrics persistence")
		metrics.MarkReady()
		return
	}

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
		return
	}

	log.Infof("Start metrics daemon, update each %v into %v", metrics.refreshRate, metrics.output)

	for {
		select {
		case <-metrics.Done():
			log.Info("Stopping metrics daemon")
			metrics.Persist()
			log.Info("Stop metrics daemon")
			return
		case <-ticker.C:
			metrics.Persist()
		}
	}
}
