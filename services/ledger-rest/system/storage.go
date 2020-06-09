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

package system

import (
	"context"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/utils"
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
		DaemonSupport: utils.NewDaemonSupport(ctx, "storage-check"),
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

// Start handles everything needed to start storage daemon
func (monitor DiskMonitor) Start() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	monitor.CheckDiskSpace()
	monitor.MarkReady()

	select {
	case <-monitor.CanStart:
		break
	case <-monitor.Done():
		monitor.MarkDone()
		return
	}

	log.Info("Start disk space monitor daemon")

	go func() {
		for {
			select {
			case <-monitor.Done():
				monitor.MarkDone()
				return
			case <-ticker.C:
				monitor.CheckDiskSpace()
			}
		}
	}()

	monitor.WaitStop()
	log.Info("Stop disk space monitor daemon")
}
