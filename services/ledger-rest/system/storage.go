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
	"github.com/jancajthaml-openbank/ledger-rest/utils"
	"sync/atomic"
	"syscall"
	"time"
)

// DiskMonitor represents disk monitoring subroutine
type DiskMonitor struct {
	utils.DaemonSupport
	rootStorage string
	limit       uint64
	free        uint64
	used        uint64
	ok          int32
}

// NewDiskMonitor returns new disk monitor fascade
func NewDiskMonitor(ctx context.Context, limit uint64, rootStorage string) *DiskMonitor {
	return &DiskMonitor{
		DaemonSupport: utils.NewDaemonSupport(ctx, "storage-check"),
		rootStorage:   rootStorage,
		limit:         limit,
		free:          0,
		used:          0,
		ok:            1,
	}
}

// IsHealthy true if storage is healthy
func (monitor *DiskMonitor) IsHealthy() bool {
	if monitor == nil {
		return true
	}
	return atomic.LoadInt32(&(monitor.ok)) != 0
}

// GetFreeDiskSpace returns free disk space
func (monitor *DiskMonitor) GetFreeDiskSpace() uint64 {
	if monitor == nil {
		return 0
	}
	return atomic.LoadUint64(&(monitor.free))
}

// GetUsedDiskSpace returns used disk space
func (monitor *DiskMonitor) GetUsedDiskSpace() uint64 {
	if monitor == nil {
		return 0
	}
	return atomic.LoadUint64(&(monitor.used))
}

// CheckDiskSpace update free disk space metric and determine if ok to operate
func (monitor *DiskMonitor) CheckDiskSpace() {
	if monitor == nil {
		return
	}
	var stat = new(syscall.Statfs_t)
	err := syscall.Statfs(monitor.rootStorage, stat)
	if err != nil {
		log.Warn().Msgf("Unable to obtain storage stats")
		atomic.StoreInt32(&(monitor.ok), 0)
		return
	}
	free := stat.Bavail * uint64(stat.Bsize)
	used := (stat.Blocks - stat.Bfree) * uint64(stat.Bsize)

	atomic.StoreUint64(&(monitor.free), free)
	atomic.StoreUint64(&(monitor.used), used)

	if monitor.limit > 0 && free < monitor.limit {
		log.Warn().Msg("Not enough disk space to continue operating")
		atomic.StoreInt32(&(monitor.ok), 0)
		return
	}
	atomic.StoreInt32(&(monitor.ok), 1)
	return
}

// Start handles everything needed to start storage daemon
func (monitor *DiskMonitor) Start() {
	if monitor == nil {
		return
	}
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

	log.Info().Msg("Start disk space monitor daemon")

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
	log.Info().Msg("Stop disk space monitor daemon")
}
