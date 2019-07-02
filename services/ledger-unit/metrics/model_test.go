package metrics

import (
	"testing"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		_, err := entity.MarshalJSON()
		assert.EqualError(t, err, "cannot marshall nil")
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		_, err := entity.MarshalJSON()
		assert.EqualError(t, err, "cannot marshall nil references")
	}

	t.Log("happy path")
	{
		entity := Metrics{
			promisedTransactions:   metrics.NewCounter(),
			promisedTransfers:      metrics.NewCounter(),
			committedTransactions:  metrics.NewCounter(),
			committedTransfers:     metrics.NewCounter(),
			rollbackedTransactions: metrics.NewCounter(),
			rollbackedTransfers:    metrics.NewCounter(),
			forwardedTransactions:  metrics.NewCounter(),
			forwardedTransfers:     metrics.NewCounter(),
		}

		entity.promisedTransactions.Inc(1)
		entity.promisedTransfers.Inc(2)
		entity.committedTransactions.Inc(3)
		entity.committedTransfers.Inc(4)
		entity.rollbackedTransactions.Inc(5)
		entity.rollbackedTransfers.Inc(6)
		entity.forwardedTransactions.Inc(7)
		entity.forwardedTransfers.Inc(8)

		actual, err := entity.MarshalJSON()

		require.Nil(t, err)

		data := []byte("{\"promisedTransactions\":1,\"promisedTransfers\":2,\"committedTransactions\":3,\"committedTransfers\":4,\"rollbackedTransactions\":5,\"rollbackedTransfers\":6,\"forwardedTransactions\":7,\"forwardedTransfers\":8}")

		assert.Equal(t, data, actual)
	}
}

func TestUnmarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		err := entity.UnmarshalJSON([]byte(""))
		assert.EqualError(t, err, "cannot unmarshall to nil")
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		err := entity.UnmarshalJSON([]byte(""))
		assert.EqualError(t, err, "cannot unmarshall to nil references")
	}

	t.Log("error on malformed data")
	{
		entity := Metrics{
			promisedTransactions:   metrics.NewCounter(),
			promisedTransfers:      metrics.NewCounter(),
			committedTransactions:  metrics.NewCounter(),
			committedTransfers:     metrics.NewCounter(),
			rollbackedTransactions: metrics.NewCounter(),
			rollbackedTransfers:    metrics.NewCounter(),
			forwardedTransactions:  metrics.NewCounter(),
			forwardedTransfers:     metrics.NewCounter(),
		}

		data := []byte("{")
		assert.NotNil(t, entity.UnmarshalJSON(data))
	}

	t.Log("happy path")
	{
		entity := Metrics{
			promisedTransactions:   metrics.NewCounter(),
			promisedTransfers:      metrics.NewCounter(),
			committedTransactions:  metrics.NewCounter(),
			committedTransfers:     metrics.NewCounter(),
			rollbackedTransactions: metrics.NewCounter(),
			rollbackedTransfers:    metrics.NewCounter(),
			forwardedTransactions:  metrics.NewCounter(),
			forwardedTransfers:     metrics.NewCounter(),
		}

		data := []byte("{\"promisedTransactions\":1,\"promisedTransfers\":2,\"committedTransactions\":3,\"committedTransfers\":4,\"rollbackedTransactions\":5,\"rollbackedTransfers\":6,\"forwardedTransactions\":7,\"forwardedTransfers\":8}")
		require.Nil(t, entity.UnmarshalJSON(data))

		assert.Equal(t, int64(1), entity.promisedTransactions.Count())
		assert.Equal(t, int64(2), entity.promisedTransfers.Count())
		assert.Equal(t, int64(3), entity.committedTransactions.Count())
		assert.Equal(t, int64(4), entity.committedTransfers.Count())
		assert.Equal(t, int64(5), entity.rollbackedTransactions.Count())
		assert.Equal(t, int64(6), entity.rollbackedTransfers.Count())
		assert.Equal(t, int64(7), entity.forwardedTransactions.Count())
		assert.Equal(t, int64(8), entity.forwardedTransfers.Count())

	}
}
