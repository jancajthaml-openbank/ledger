package system

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiskIsHealthy(t *testing.T) {
	t.Log("true if caller is nil")
	{
		var monitor *DiskMonitor
		assert.Equal(t, true, monitor.IsHealthy())
	}

	t.Log("true by default")
	{
		monitor := NewDiskMonitor(uint64(0), "/tmp")
		assert.Equal(t, true, monitor.IsHealthy())
	}

	t.Log("false is attribute ok=0")
	{
		monitor := NewDiskMonitor(uint64(0), "/tmp")
		atomic.StoreInt32(&(monitor.ok), 0)
		assert.Equal(t, false, monitor.IsHealthy())
	}
}

func TestGetDiskFree(t *testing.T) {
	t.Log("0 if caller is nil")
	{
		var monitor *DiskMonitor
		assert.Equal(t, uint64(0), monitor.GetFree())
	}

	t.Log("0 by default")
	{
		monitor := NewDiskMonitor(uint64(0), "/tmp")
		assert.Equal(t, uint64(0), monitor.GetFree())
	}

	t.Log("10 is attribute free=10")
	{
		monitor := NewDiskMonitor(uint64(0), "/tmp")
		atomic.StoreUint64(&(monitor.free), 10)
		assert.Equal(t, uint64(10), monitor.GetFree())
	}
}

func TestGetDiskUsed(t *testing.T) {
	t.Log("0 if caller is nil")
	{
		var monitor *DiskMonitor
		assert.Equal(t, uint64(0), monitor.GetUsed())
	}

	t.Log("0 by default")
	{
		monitor := NewDiskMonitor(uint64(0), "/tmp")
		assert.Equal(t, uint64(0), monitor.GetUsed())
	}

	t.Log("10 is attribute used=10")
	{
		monitor := NewDiskMonitor(uint64(0), "/tmp")
		atomic.StoreUint64(&(monitor.used), 10)
		assert.Equal(t, uint64(10), monitor.GetUsed())
	}
}

func TestCheckDiskSpace(t *testing.T) {
	t.Log("does not panic if caller is nil")
	{
		var monitor *DiskMonitor
		monitor.CheckDiskSpace()
	}

	t.Log("ok=1 if available memory is above limit")
	{
		monitor := NewDiskMonitor(uint64(1), "/tmp")
		monitor.CheckDiskSpace()
		assert.Equal(t, true, monitor.IsHealthy())
	}

	t.Log("ok=0 if available memory is under limit")
	{
		monitor := NewDiskMonitor(^uint64(0), "/tmp")
		monitor.CheckDiskSpace()
		assert.Equal(t, false, monitor.IsHealthy())
	}
}
