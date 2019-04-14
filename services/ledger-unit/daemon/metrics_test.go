package daemon

import (
	"context"
	"testing"

	"github.com/jancajthaml-openbank/ledger-unit/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFilename(t *testing.T) {
	assert.Equal(t, "/a/b/c.d.e", getFilename("/a/b/c.e", "d"))
	assert.Equal(t, "/a/b/c.d", getFilename("/a/b/c.d", ""))
}

func TestMetricsPersist(t *testing.T) {
	cfg := config.Configuration{}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	entity := NewMetrics(ctx, cfg)

	t.Log("TransactionPromised properly increments number of promised transactions and transfers")
	{
		require.Equal(t, int64(0), entity.promisedTransactions.Count())
		require.Equal(t, int64(0), entity.promisedTransfers.Count())
		entity.TransactionPromised(2)
		assert.Equal(t, int64(1), entity.promisedTransactions.Count())
		assert.Equal(t, int64(2), entity.promisedTransfers.Count())
	}

	t.Log("TransactionCommitted properly increments number of committed transactions and transfers")
	{
		require.Equal(t, int64(0), entity.committedTransactions.Count())
		require.Equal(t, int64(0), entity.committedTransfers.Count())
		entity.TransactionCommitted(2)
		assert.Equal(t, int64(1), entity.committedTransactions.Count())
		assert.Equal(t, int64(2), entity.committedTransfers.Count())
	}

	t.Log("TransactionRollbacked properly increments number of rollbacked transactions and transfers")
	{
		require.Equal(t, int64(0), entity.rollbackedTransactions.Count())
		require.Equal(t, int64(0), entity.rollbackedTransfers.Count())
		entity.TransactionRollbacked(2)
		assert.Equal(t, int64(1), entity.rollbackedTransactions.Count())
		assert.Equal(t, int64(2), entity.rollbackedTransfers.Count())
	}

	t.Log("TransactionForwarded properly increments number of forwarded transactions and transfers")
	{
		require.Equal(t, int64(0), entity.forwardedTransactions.Count())
		require.Equal(t, int64(0), entity.forwardedTransfers.Count())
		entity.TransactionForwarded(2)
		assert.Equal(t, int64(1), entity.forwardedTransactions.Count())
		assert.Equal(t, int64(2), entity.forwardedTransfers.Count())
	}
}
