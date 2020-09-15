package metrics

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	localfs "github.com/jancajthaml-openbank/local-fs"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		_, err := entity.MarshalJSON()
		assert.NotNil(t, err)
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		_, err := entity.MarshalJSON()
		assert.NotNil(t, err)
	}

	t.Log("happy path")
	{
		entity := Metrics{
			createTransactionLatency: metrics.NewTimer(),
		}

		entity.createTransactionLatency.Update(time.Duration(3))

		actual, err := entity.MarshalJSON()

		require.Nil(t, err)

		aux := &struct {
			CreateTransactionLatency float64 `json:"createTransactionLatency"`
		}{}

		require.Nil(t, json.Unmarshal(actual, &aux))

		assert.Equal(t, float64(3), aux.CreateTransactionLatency)
	}
}

func TestUnmarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		err := entity.UnmarshalJSON([]byte(""))
		assert.NotNil(t, err)
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		err := entity.UnmarshalJSON([]byte(""))
		assert.NotNil(t, err)
	}

	t.Log("error on malformed data")
	{
		entity := Metrics{
			createTransactionLatency: metrics.NewTimer(),
		}

		data := []byte("{")
		assert.NotNil(t, entity.UnmarshalJSON(data))
	}

	t.Log("happy path")
	{
		entity := Metrics{
			createTransactionLatency: metrics.NewTimer(),
		}

		data := []byte("{\"createTransactionLatency\":3}")
		require.Nil(t, entity.UnmarshalJSON(data))

		assert.Equal(t, float64(3), entity.createTransactionLatency.Percentile(0.95))
	}
}

func TestPersist(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		assert.EqualError(t, entity.Persist(), "cannot persist nil reference")
	}

	t.Log("error when marshaling fails")
	{
		entity := Metrics{}
		assert.NotNil(t, entity.Persist())
	}

	t.Log("happy path")
	{
		defer os.Remove("/tmp/metrics.json")

		entity := Metrics{
			storage:                  localfs.NewPlaintextStorage("/tmp"),
			createTransactionLatency: metrics.NewTimer(),
		}

		require.Nil(t, entity.Persist())

		expected, err := entity.MarshalJSON()
		require.Nil(t, err)

		actual, err := ioutil.ReadFile("/tmp/metrics.json")
		require.Nil(t, err)

		assert.Equal(t, expected, actual)
	}
}

func TestHydrate(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		assert.EqualError(t, entity.Hydrate(), "cannot hydrate nil reference")
	}

	t.Log("happy path")
	{
		defer os.Remove("/tmp/metrics.json")

		old := Metrics{
			createTransactionLatency: metrics.NewTimer(),
		}

		old.createTransactionLatency.Update(time.Duration(3))

		data, err := old.MarshalJSON()
		require.Nil(t, err)

		require.Nil(t, ioutil.WriteFile("/tmp/metrics.json", data, 0444))

		entity := Metrics{
			storage:                  localfs.NewPlaintextStorage("/tmp"),
			createTransactionLatency: metrics.NewTimer(),
		}

		require.Nil(t, entity.Hydrate())

		assert.Equal(t, float64(3), entity.createTransactionLatency.Percentile(0.95))
	}
}
