package metrics

import (
	"testing"
	"time"

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
			getTransactionLatency:    metrics.NewTimer(),
			getTransactionsLatency:   metrics.NewTimer(),
			createTransactionLatency: metrics.NewTimer(),
			forwardTransferLatency:   metrics.NewTimer(),
		}

		entity.getTransactionLatency.Update(time.Duration(1))
		entity.getTransactionsLatency.Update(time.Duration(2))
		entity.createTransactionLatency.Update(time.Duration(3))
		entity.forwardTransferLatency.Update(time.Duration(4))

		actual, err := entity.MarshalJSON()

		require.Nil(t, err)

		data := []byte("{\"getTransactionLatency\":1,\"getTransactionsLatency\":2,\"createTransactionLatency\":3,\"forwardTransferLatency\":4}")

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
			getTransactionLatency:    metrics.NewTimer(),
			getTransactionsLatency:   metrics.NewTimer(),
			createTransactionLatency: metrics.NewTimer(),
			forwardTransferLatency:   metrics.NewTimer(),
		}

		data := []byte("{")
		assert.NotNil(t, entity.UnmarshalJSON(data))
	}

	t.Log("happy path")
	{
		entity := Metrics{
			getTransactionLatency:    metrics.NewTimer(),
			getTransactionsLatency:   metrics.NewTimer(),
			createTransactionLatency: metrics.NewTimer(),
			forwardTransferLatency:   metrics.NewTimer(),
		}

		data := []byte("{\"getTransactionLatency\":1,\"getTransactionsLatency\":2,\"createTransactionLatency\":3,\"forwardTransferLatency\":4}")
		require.Nil(t, entity.UnmarshalJSON(data))

		assert.Equal(t, float64(1), entity.getTransactionLatency.Percentile(0.95))
		assert.Equal(t, float64(2), entity.getTransactionsLatency.Percentile(0.95))
		assert.Equal(t, float64(3), entity.createTransactionLatency.Percentile(0.95))
		assert.Equal(t, float64(4), entity.forwardTransferLatency.Percentile(0.95))

	}
}
