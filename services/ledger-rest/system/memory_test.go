package system

import (
	"time"
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryIsHealthy(t *testing.T) {
	t.Log("true if caller is nil")
	{
		var monitor *MemoryMonitor
		assert.Equal(t, true, monitor.IsHealthy())
	}

	t.Log("true by default")
	{
		monitor := NewMemoryMonitor(context.Background(), uint64(0))
		assert.Equal(t, true, monitor.IsHealthy())
	}

	t.Log("false is attribute ok=0")
	{
		monitor := NewMemoryMonitor(context.Background(), uint64(0))
		atomic.StoreInt32(&(monitor.ok), 0)
		assert.Equal(t, false, monitor.IsHealthy())
	}
}

func TestGetMemoryFree(t *testing.T) {
	t.Log("0 if caller is nil")
	{
		var monitor *MemoryMonitor
		assert.Equal(t, uint64(0), monitor.GetFree())
	}

	t.Log("0 by default")
	{
		monitor := NewMemoryMonitor(context.Background(), uint64(0))
		assert.Equal(t, uint64(0), monitor.GetFree())
	}

	t.Log("10 is attribute free=10")
	{
		monitor := NewMemoryMonitor(context.Background(), uint64(0))
		atomic.StoreUint64(&(monitor.free), 10)
		assert.Equal(t, uint64(10), monitor.GetFree())
	}
}

func TestGetMemoryUsed(t *testing.T) {
	t.Log("0 if caller is nil")
	{
		var monitor *MemoryMonitor
		assert.Equal(t, uint64(0), monitor.GetUsed())
	}

	t.Log("0 by default")
	{
		monitor := NewMemoryMonitor(context.Background(), uint64(0))
		assert.Equal(t, uint64(0), monitor.GetUsed())
	}

	t.Log("10 is attribute used=10")
	{
		monitor := NewMemoryMonitor(context.Background(), uint64(0))
		atomic.StoreUint64(&(monitor.used), 10)
		assert.Equal(t, uint64(10), monitor.GetUsed())
	}
}

func TestCheckMemoryAllocation(t *testing.T) {
	t.Log("does not panic if caller is nil")
	{
		var monitor *MemoryMonitor
		monitor.CheckMemoryAllocation()
	}

	t.Log("ok=1 if available memory is above limit")
	{
		monitor := NewMemoryMonitor(context.Background(), uint64(1))
		monitor.CheckMemoryAllocation()
		assert.Equal(t, true, monitor.IsHealthy())
	}

	t.Log("ok=0 if available memory is under limit")
	{
		monitor := NewMemoryMonitor(context.Background(), ^uint64(0))
		monitor.CheckMemoryAllocation()
		assert.Equal(t, false, monitor.IsHealthy())
	}
}

func TestMemoryMonitorDaemonSupport(t *testing.T) {

	t.Log("does not panic if nil")
	{
		var monitor *MemoryMonitor
		monitor.Start()
		// FIXME panics
		//monitor.Stop()
	}

	t.Log("parent context canceled before even started")
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		cancel()

		monitor := NewMemoryMonitor(ctx, uint64(0))

		go monitor.Start()
		<-monitor.IsReady
		monitor.GreenLight()
		monitor.WaitStop()
	}

	t.Log("parent context canceled while already running")
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		monitor := NewMemoryMonitor(ctx, uint64(0))

		go monitor.Start()
		<-monitor.IsReady
		monitor.GreenLight()
		cancel()
		monitor.WaitStop()
	}

	t.Log("manual Start -> Stop")
	{
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		monitor := NewMemoryMonitor(ctx, uint64(0))

		go monitor.Start()
		<-monitor.IsReady
		monitor.GreenLight()
		monitor.Stop()
		monitor.WaitStop()
	}
}
