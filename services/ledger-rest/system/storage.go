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

package system

import (
	"context"
	"fmt"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/utils"

	log "github.com/sirupsen/logrus"
)

// DiskMonitor represents disk monitoring subroutine
type DiskMonitor struct {
	utils.DaemonSupport
	rootStorage string
	limit       uint64
	free        *uint64
	used        *uint64
	ok          *int32
}

// NewDiskMonitor returns new disk monitor fascade
func NewDiskMonitor(ctx context.Context, limit uint64, rootStorage string) DiskMonitor {
	ok := int32(1)
	free := uint64(0)
	used := uint64(0)
	return DiskMonitor{
		DaemonSupport: utils.NewDaemonSupport(ctx),
		rootStorage:   rootStorage,
		limit:         limit,
		free:          &free,
		used:          &used,
		ok:            &ok,
	}
}

// IsHealthy true if storage is healthy
func (monitor *DiskMonitor) IsHealthy() bool {
	return atomic.LoadInt32(monitor.ok) != 0
}

// GetFreeDiskSpace returns free disk space
func (monitor *DiskMonitor) GetFreeDiskSpace() uint64 {
	return atomic.LoadUint64(monitor.free)
}

// GetUsedDiskSpace returns used disk space
func (monitor *DiskMonitor) GetUsedDiskSpace() uint64 {
	return atomic.LoadUint64(monitor.used)
}

// CheckDiskSpace update free disk space metric and determine if ok to operate
func (monitor *DiskMonitor) CheckDiskSpace() {
	defer recover()

	var stat = new(syscall.Statfs_t)
	syscall.Statfs(monitor.rootStorage, stat)

	free := stat.Bavail * uint64(stat.Bsize)
	used := (stat.Blocks - stat.Bfree) * uint64(stat.Bsize)

	atomic.StoreUint64(monitor.free, free)
	atomic.StoreUint64(monitor.used, used)

	if monitor.limit > 0 && free < monitor.limit {
		log.Warnf("Not enough disk space to continue operating")
		atomic.StoreInt32(monitor.ok, 0)
		return
	}
	atomic.StoreInt32(monitor.ok, 1)
	return
}

// WaitReady wait for daemon to be ready
func (monitor DiskMonitor) WaitReady(deadline time.Duration) (err error) {
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
	case <-monitor.IsReady:
		ticker.Stop()
		err = nil
		return
	case <-ticker.C:
		err = fmt.Errorf("daemon was not ready within %v seconds", deadline)
		return
	}
}

// Start handles everything needed to start daemon
func (monitor DiskMonitor) Start() {
	defer monitor.MarkDone()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	monitor.CheckDiskSpace()
	monitor.MarkReady()

	select {
	case <-monitor.CanStart:
		break
	case <-monitor.Done():
		return
	}

	log.Info("Start disk space monitor daemon")

	for {
		select {
		case <-monitor.Done():
			log.Info("Stopping disk space monitor daemon")
			log.Info("Stop disk space monitor daemon")
			return
		case <-ticker.C:
			monitor.CheckDiskSpace()
		}
	}
}
