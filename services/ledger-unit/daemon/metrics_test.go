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

	t.Log("TransactionPromised properly increments number of created accounts")
	{
		require.Equal(t, int64(0), entity.promisedTransactions.Count())
		entity.TransactionPromised()
		assert.Equal(t, int64(1), entity.promisedTransactions.Count())
	}

	t.Log("TransactionCommitted properly increments number of created accounts")
	{
		require.Equal(t, int64(0), entity.committedTransactions.Count())
		entity.TransactionCommitted()
		assert.Equal(t, int64(1), entity.committedTransactions.Count())
	}

	t.Log("TransactionRollbacked properly increments number of created accounts")
	{
		require.Equal(t, int64(0), entity.rollbackedTransactions.Count())
		entity.TransactionRollbacked()
		assert.Equal(t, int64(1), entity.rollbackedTransactions.Count())
	}

	t.Log("TransactionForwarded properly increments number of created accounts")
	{
		require.Equal(t, int64(0), entity.forwardedTransactions.Count())
		entity.TransactionForwarded()
		assert.Equal(t, int64(1), entity.forwardedTransactions.Count())
	}
}
