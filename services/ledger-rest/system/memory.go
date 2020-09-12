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
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/utils"
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
		DaemonSupport: utils.NewDaemonSupport(ctx, "memory-check"),
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
		log.Warn().Msgf("Unable to obtain memory stats")
		atomic.StoreInt32(monitor.ok, 0)
		return
	}

	free := uint64(sysStat.Freeram) * uint64(sysStat.Unit)

	atomic.StoreUint64(monitor.free, free)
	atomic.StoreUint64(monitor.used, memStat.Sys)

	if monitor.limit > 0 && free < monitor.limit {
		log.Warn().Msgf("Not enough memory to continue operating")
		atomic.StoreInt32(monitor.ok, 0)
		return
	}
	atomic.StoreInt32(monitor.ok, 1)
	return
}

// Start handles everything needed to start memory daemon
func (monitor MemoryMonitor) Start() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	monitor.CheckMemoryAllocation()
	monitor.MarkReady()

	select {
	case <-monitor.CanStart:
		break
	case <-monitor.Done():
		monitor.MarkDone()
		return
	}

	log.Info().Msg("Start memory-monitor daemon")

	go func() {
		for {
			select {
			case <-monitor.Done():
				monitor.MarkDone()
				return
			case <-ticker.C:
				monitor.CheckMemoryAllocation()
			}
		}
	}()

	monitor.WaitStop()
	log.Info().Msg("Stop memory-monitor daemon")
}
