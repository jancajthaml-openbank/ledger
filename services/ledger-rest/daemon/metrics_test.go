package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFilename(t *testing.T) {
	assert.Equal(t, "/a/b/c.e", getFilename("/a/b/c.e"))
}

func TestMetricsPersist(t *testing.T) {
	cfg := config.Configuration{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	entity := NewMetrics(ctx, cfg)
	delay := 1e8
	delta := 1e8

	t.Log("TimeGetTransaction properly times run of GetAccount function")
	{
		require.Equal(t, int64(0), entity.getTransactionLatency.Count())
		entity.TimeGetTransaction(func() {
			time.Sleep(time.Duration(delay))
		})
		assert.Equal(t, int64(1), entity.getTransactionLatency.Count())
		assert.InDelta(t, entity.getTransactionLatency.Percentile(0.95), delay, delta)
	}

	t.Log("TimeGetTransactions properly times run of GetAccount function")
	{
		require.Equal(t, int64(0), entity.getTransactionsLatency.Count())
		entity.TimeGetTransactions(func() {
			time.Sleep(time.Duration(delay))
		})
		assert.Equal(t, int64(1), entity.getTransactionsLatency.Count())
		assert.InDelta(t, entity.getTransactionsLatency.Percentile(0.95), delay, delta)
	}

	t.Log("TimeCreateTransaction properly times run of CreateAccount function")
	{
		require.Equal(t, int64(0), entity.createTransactionLatency.Count())
		entity.TimeCreateTransaction(func() {
			time.Sleep(time.Duration(delay))
		})
		assert.Equal(t, int64(1), entity.createTransactionLatency.Count())
		assert.InDelta(t, entity.createTransactionLatency.Percentile(0.95), delay, delta)
	}

	t.Log("TimeForwardTransfer properly times run of CreateAccount function")
	{
		require.Equal(t, int64(0), entity.forwardTransferLatency.Count())
		entity.TimeForwardTransfer(func() {
			time.Sleep(time.Duration(delay))
		})
		assert.Equal(t, int64(1), entity.forwardTransferLatency.Count())
		assert.InDelta(t, entity.forwardTransferLatency.Percentile(0.95), delay, delta)
	}
}
