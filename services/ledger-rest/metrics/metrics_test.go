package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	entity := NewMetrics(ctx, "/tmp", time.Hour)
	delay := 1e8
	delta := 1e8

	t.Log("TimeCreateTransaction properly times run of CreateAccount function")
	{
		require.Equal(t, int64(0), entity.createTransactionLatency.Count())
		entity.TimeCreateTransaction(func() {
			time.Sleep(time.Duration(delay))
		})
		assert.Equal(t, int64(1), entity.createTransactionLatency.Count())
		assert.InDelta(t, entity.createTransactionLatency.Percentile(0.95), delay, delta)
	}
}
