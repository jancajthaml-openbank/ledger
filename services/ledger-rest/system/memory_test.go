package system

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryMonitorIsHealthy(t *testing.T) {
	t.Log("true if caller is nil")
	{
		var monitor *MemoryMonitor
		assert.Equal(t, true, monitor.IsHealthy())
	}

	t.Log("true by default")
	{
		monitor := NewMemoryMonitor(context.Background(), ^uint64(0))
		assert.Equal(t, true, monitor.IsHealthy())
	}

	t.Log("false is attribute ok=0")
	{
		monitor := NewMemoryMonitor(context.Background(), ^uint64(0))
		atomic.StoreInt32(&(monitor.ok), 0)
		assert.Equal(t, false, monitor.IsHealthy())
	}
}

func TestMemoryMonitorGetFree(t *testing.T) {
	t.Log("0 if caller is nil")
	{
		var monitor *MemoryMonitor
		assert.Equal(t, uint64(0), monitor.GetFree())
	}

	t.Log("0 by default")
	{
		monitor := NewMemoryMonitor(context.Background(), ^uint64(0))
		assert.Equal(t, uint64(0), monitor.GetFree())
	}

	t.Log("10 is attribute free=10")
	{
		monitor := NewMemoryMonitor(context.Background(), ^uint64(0))
		atomic.StoreUint64(&(monitor.free), 10)
		assert.Equal(t, uint64(10), monitor.GetFree())
	}
}

func TestMemoryMonitorGetUsed(t *testing.T) {
	t.Log("0 if caller is nil")
	{
		var monitor *MemoryMonitor
		assert.Equal(t, uint64(0), monitor.GetUsed())
	}

	t.Log("0 by default")
	{
		monitor := NewMemoryMonitor(context.Background(), ^uint64(0))
		assert.Equal(t, uint64(0), monitor.GetUsed())
	}

	t.Log("10 is attribute used=10")
	{
		monitor := NewMemoryMonitor(context.Background(), ^uint64(0))
		atomic.StoreUint64(&(monitor.used), 10)
		assert.Equal(t, uint64(10), monitor.GetUsed())
	}
}
