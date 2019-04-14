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
	"time"

	"github.com/jancajthaml-openbank/ledger-unit/config"

	system "github.com/jancajthaml-openbank/actor-system"
	localfs "github.com/jancajthaml-openbank/local-fs"
	log "github.com/sirupsen/logrus"
)

// TransactionFinalizer represents journal saturation update subroutine
type TransactionFinalizer struct {
	Support
	callback     func(msg interface{}, to system.Coordinates, from system.Coordinates)
	metrics      *Metrics
	storage      *localfs.Storage
	scanInterval time.Duration
}

// NewTransactionFinalizer returns snapshot updater fascade
func NewTransactionFinalizer(ctx context.Context, cfg config.Configuration, metrics *Metrics, storage *localfs.Storage, callback func(msg interface{}, to system.Coordinates, from system.Coordinates)) TransactionFinalizer {
	return TransactionFinalizer{
		Support:      NewDaemonSupport(ctx),
		callback:     callback,
		metrics:      metrics,
		storage:      storage,
		scanInterval: cfg.TransactionIntegrityScanInterval,
	}
}

func (scan TransactionFinalizer) performIntegrityScan() {

}

/*
// FIXME unit test coverage
// FIXME maximum events to params
func (updater SnapshotUpdater) updateSaturated() {
	accounts := updater.getAccounts()
	var numberOfSnapshotsUpdated int64

	for _, name := range accounts {
		version := updater.getVersion(name)
		if version == -1 {
			continue
		}
		if updater.getEvents(name, version) >= updater.saturationThreshold {
			log.Debugf("Request %v to update snapshot version from %d to %d", name, version, version+1)
			msg := model.Update{Version: version}
			to := system.Coordinates{Name: name}
			from := system.Coordinates{Name: "snapshot_saturation_cron"}
			updater.callback(msg, to, from)

			numberOfSnapshotsUpdated++
		}
	}
	updater.metrics.SnapshotsUpdated(numberOfSnapshotsUpdated)
}

func (updater SnapshotUpdater) getAccounts() []string {
	result, err := updater.storage.ListDirectory(utils.RootPath(), true)
	if err != nil {
		return nil
	}
	return result
}

func (updater SnapshotUpdater) getVersion(name string) int {
	result, err := updater.storage.ListDirectory(utils.SnapshotsPath(name), false)
	if err != nil || len(result) == 0 {
		return -1
	}

	version, err := strconv.Atoi(result[0])
	if err != nil {
		return -1
	}

	return version
}

func (updater SnapshotUpdater) getEvents(name string, version int) int {
	result, err := updater.storage.CountFiles(utils.EventPath(name, version))
	if err != nil {
		return -1
	}
	return result
}*/

// WaitReady wait for snapshot updated to be ready
func (scan TransactionFinalizer) WaitReady(deadline time.Duration) (err error) {
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
	case <-scan.IsReady:
		ticker.Stop()
		err = nil
		return
	case <-ticker.C:
		err = fmt.Errorf("daemon was not ready within %v seconds", deadline)
		return
	}
}

// Start handles everything needed to start snapshot updater daemon it runs scan
// of accounts snapshots and events and orders accounts to update their snapshot
// if number of events in given version is larger than threshold
func (scan TransactionFinalizer) Start() {
	defer scan.MarkDone()

	ticker := time.NewTicker(scan.scanInterval)
	defer ticker.Stop()

	scan.MarkReady()

	select {
	case <-scan.canStart:
		break
	case <-scan.Done():
		return
	}

	log.Infof("Start transaction integrity check daemon, scan each %v", scan.scanInterval)

	for {
		select {
		case <-scan.Done():
			log.Info("Stopping transaction integrity check daemon")
			log.Info("Stop transaction integrity check daemon")
			return
		case <-ticker.C:
			//updater.metrics.TimeUpdateSaturatedSnapshots(func() {
			scan.performIntegrityScan()
			//})
		}
	}
}
