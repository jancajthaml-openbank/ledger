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

package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jancajthaml-openbank/ledger-unit/config"
	"github.com/jancajthaml-openbank/ledger-unit/utils"

	metrics "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

// Metrics represents metrics subroutine
type Metrics struct {
	Support
	output                 string
	tenant                 string
	refreshRate            time.Duration
	promisedTransactions   metrics.Counter
	committedTransactions  metrics.Counter
	rollbackedTransactions metrics.Counter
	forwardedTransactions  metrics.Counter
}

// NewMetrics returns metrics fascade
func NewMetrics(ctx context.Context, cfg config.Configuration) Metrics {
	return Metrics{
		Support:                NewDaemonSupport(ctx),
		output:                 cfg.MetricsOutput,
		tenant:                 cfg.Tenant,
		refreshRate:            cfg.MetricsRefreshRate,
		promisedTransactions:   metrics.NewCounter(),
		committedTransactions:  metrics.NewCounter(),
		rollbackedTransactions: metrics.NewCounter(),
		forwardedTransactions:  metrics.NewCounter(),
	}
}

// Snapshot holds metrics snapshot status
type Snapshot struct {
	PromisedTransactions   int64 `json:"promisedTransactions"`
	CommittedTransactions  int64 `json:"committedTransactions"`
	RollbackedTransactions int64 `json:"rollbackedTransactions"`
	ForwardedTransactions  int64 `json:"forwardedTransactions"`
}

// NewSnapshot returns metrics snapshot
func NewSnapshot(metrics Metrics) Snapshot {
	return Snapshot{
		PromisedTransactions:   metrics.promisedTransactions.Count(),
		CommittedTransactions:  metrics.committedTransactions.Count(),
		RollbackedTransactions: metrics.rollbackedTransactions.Count(),
		ForwardedTransactions:  metrics.forwardedTransactions.Count(),
	}
}

// TransactionPromised increments transactions promised by one
func (metrics Metrics) TransactionPromised() {
	metrics.promisedTransactions.Inc(1)
}

// TransactionCommitted increments transactions committed by one
func (metrics Metrics) TransactionCommitted() {
	metrics.committedTransactions.Inc(1)
}

// TransactionRollbacked increments transactions rollbacked by one
func (metrics Metrics) TransactionRollbacked() {
	metrics.rollbackedTransactions.Inc(1)
}

// TransactionForwarded increments transactions forwarded by one
func (metrics Metrics) TransactionForwarded() {
	metrics.forwardedTransactions.Inc(1)
}

func (metrics Metrics) persist(filename string) {
	tempFile := filename + "_temp"

	data, err := utils.JSON.Marshal(NewSnapshot(metrics))
	if err != nil {
		log.Warnf("unable to create serialize metrics with error: %v", err)
		return
	}
	f, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Warnf("unable to create file with error: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		log.Warnf("unable to write file with error: %v", err)
		return
	}

	if err := os.Rename(tempFile, filename); err != nil {
		log.Warnf("unable to move file with error: %v", err)
		return
	}

	return
}

func getFilename(path string, tenant string) string {
	if tenant == "" {
		return path
	}

	dirname := filepath.Dir(path)
	ext := filepath.Ext(path)
	filename := filepath.Base(path)
	filename = filename[:len(filename)-len(ext)]

	return dirname + "/" + filename + "." + tenant + ext
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

	output := getFilename(metrics.output, metrics.tenant)
	ticker := time.NewTicker(metrics.refreshRate)
	defer ticker.Stop()

	metrics.MarkReady()

	select {
	case <-metrics.canStart:
		break
	case <-metrics.Done():
		return
	}

	log.Infof("Start metrics daemon, update each %v into %v", metrics.refreshRate, output)

	for {
		select {
		case <-metrics.Done():
			log.Info("Stopping metrics daemon")
			metrics.persist(output)
			log.Info("Stop metrics daemon")
			return
		case <-ticker.C:
			metrics.persist(output)
		}
	}
}
