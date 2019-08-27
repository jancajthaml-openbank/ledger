package metrics

import (
	"encoding/json"
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
			createTransactionLatency: metrics.NewTimer(),
			forwardTransferLatency:   metrics.NewTimer(),
		}

		entity.createTransactionLatency.Update(time.Duration(3))
		entity.forwardTransferLatency.Update(time.Duration(4))

		actual, err := entity.MarshalJSON()

		require.Nil(t, err)

		aux := &struct {
			CreateTransactionLatency float64 `json:"createTransactionLatency"`
			ForwardTransferLatency   float64 `json:"forwardTransferLatency"`
		}{}

		require.Nil(t, json.Unmarshal(actual, &aux))

		assert.Equal(t, float64(3), aux.CreateTransactionLatency)
		assert.Equal(t, float64(4), aux.ForwardTransferLatency)
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
			createTransactionLatency: metrics.NewTimer(),
			forwardTransferLatency:   metrics.NewTimer(),
		}

		data := []byte("{")
		assert.NotNil(t, entity.UnmarshalJSON(data))
	}

	t.Log("happy path")
	{
		entity := Metrics{
			createTransactionLatency: metrics.NewTimer(),
			forwardTransferLatency:   metrics.NewTimer(),
		}

		data := []byte("{\"createTransactionLatency\":3,\"forwardTransferLatency\":4}")
		require.Nil(t, entity.UnmarshalJSON(data))

		assert.Equal(t, float64(3), entity.createTransactionLatency.Percentile(0.95))
		assert.Equal(t, float64(4), entity.forwardTransferLatency.Percentile(0.95))

	}
}
