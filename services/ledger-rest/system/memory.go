// Copyright (c) 2016-2023, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"runtime"
	"sync/atomic"
	"syscall"
)

// MemoryMonitor monitors capacity of memory
type MemoryMonitor struct {
	limit uint64
	free  uint64
	used  uint64
	ok    int32
}

// NewMemoryMonitor returns new memory monitor fascade
func NewMemoryMonitor(limit uint64) *MemoryMonitor {
	return &MemoryMonitor{
		limit: limit,
		free:  0,
		used:  0,
		ok:    1,
	}
}

// IsHealthy true if storage is healthy
func (monitor *MemoryMonitor) IsHealthy() bool {
	if monitor == nil {
		return true
	}
	return atomic.LoadInt32(&(monitor.ok)) != 0
}

// GetFree returns free memory
func (monitor *MemoryMonitor) GetFree() uint64 {
	if monitor == nil {
		return 0
	}
	return atomic.LoadUint64(&(monitor.free))
}

// GetUsed returns allocated memory
func (monitor *MemoryMonitor) GetUsed() uint64 {
	if monitor == nil {
		return 0
	}
	return atomic.LoadUint64(&(monitor.used))
}

// CheckMemoryAllocation update memory allocation metric and determine if ok to operate
func (monitor *MemoryMonitor) CheckMemoryAllocation() {
	if monitor == nil {
		return
	}

	var memStat = new(runtime.MemStats)
	runtime.ReadMemStats(memStat)

	var sysStat = new(syscall.Sysinfo_t)
	err := syscall.Sysinfo(sysStat)
	if err != nil {
		log.Warn().Msgf("Unable to obtain memory stats")
		atomic.StoreInt32(&(monitor.ok), 0)
		return
	}

	free := uint64(sysStat.Freeram) * uint64(sysStat.Unit)

	atomic.StoreUint64(&(monitor.free), free)
	atomic.StoreUint64(&(monitor.used), memStat.Sys)

	if monitor.limit > 0 && free < monitor.limit {
		log.Warn().Msgf("Not enough memory to continue operating")
		atomic.StoreInt32(&(monitor.ok), 0)
		return
	}
	atomic.StoreInt32(&(monitor.ok), 1)
	return
}

// Setup does nothing
func (monitor *MemoryMonitor) Setup() error {
	return nil
}

// Done always returns done
func (monitor *MemoryMonitor) Done() <-chan interface{} {
	done := make(chan interface{})
	close(done)
	return done
}

// Cancel does nothing
func (monitor *MemoryMonitor) Cancel() {
}

// Work checks memory allocation
func (monitor *MemoryMonitor) Work() {
	monitor.CheckMemoryAllocation()
}
