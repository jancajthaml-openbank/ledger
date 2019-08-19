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
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/utils"

	log "github.com/sirupsen/logrus"
)

// MemoryMonitor represents memory monitoring subroutine
type MemoryMonitor struct {
	utils.DaemonSupport
	limit uint64
	free  *uint64
	used  *uint64
	ok    *int32
}

// NewMemoryMonitor returns new memory monitor fascade
func NewMemoryMonitor(ctx context.Context, limit uint64) MemoryMonitor {
	ok := int32(1)
	free := uint64(0)
	used := uint64(0)

	return MemoryMonitor{
		DaemonSupport: utils.NewDaemonSupport(ctx),
		limit:         limit,
		free:          &free,
		used:          &used,
		ok:            &ok,
	}
}

// IsHealthy true if storage is healthy
func (monitor *MemoryMonitor) IsHealthy() bool {
	return atomic.LoadInt32(monitor.ok) != 0
}

// GetFreeMemory returns free memory
func (monitor *MemoryMonitor) GetFreeMemory() uint64 {
	return atomic.LoadUint64(monitor.free)
}

// GetUsedMemory returns allocated memory
func (monitor *MemoryMonitor) GetUsedMemory() uint64 {
	return atomic.LoadUint64(monitor.used)
}

// CheckMemoryAllocation update memory allocation metric and determine if ok to operate
func (monitor *MemoryMonitor) CheckMemoryAllocation() {
	defer recover()

	var memStat runtime.MemStats
	runtime.ReadMemStats(&memStat)

	var sysStat = new(syscall.Sysinfo_t)
	err := syscall.Sysinfo(sysStat)
	if err != nil {
		log.Warnf("Unable to obtain memory stats")
		atomic.StoreInt32(monitor.ok, 0)
		return
	}

	free := uint64(sysStat.Freeram) * uint64(sysStat.Unit)

	atomic.StoreUint64(monitor.free, free)
	atomic.StoreUint64(monitor.used, memStat.Sys)

	if monitor.limit > 0 && free < monitor.limit {
		log.Warnf("Not enough memory to continue operating")
		atomic.StoreInt32(monitor.ok, 0)
		return
	}
	atomic.StoreInt32(monitor.ok, 1)
	return
}

// WaitReady wait for daemon to be ready
func (monitor MemoryMonitor) WaitReady(deadline time.Duration) (err error) {
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
func (monitor MemoryMonitor) Start() {
	defer monitor.MarkDone()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	monitor.CheckMemoryAllocation()
	monitor.MarkReady()

	select {
	case <-monitor.CanStart:
		break
	case <-monitor.Done():
		return
	}

	log.Info("Start memory monitor daemon")

	for {
		select {
		case <-monitor.Done():
			log.Info("Stopping memory monitor daemon")
			log.Info("Stop memory monitor daemon")
			return
		case <-ticker.C:
			monitor.CheckMemoryAllocation()
		}
	}
}
