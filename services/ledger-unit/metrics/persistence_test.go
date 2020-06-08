package metrics

import (
	"io/ioutil"
	"os"
	"time"
	"testing"

	localfs "github.com/jancajthaml-openbank/local-fs"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersist(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		assert.EqualError(t, entity.Persist(), "cannot persist nil reference")
	}

	t.Log("error when marshalling fails")
	{
		entity := Metrics{}
		assert.EqualError(t, entity.Persist(), "cannot marshall nil references")
	}

	t.Log("happy path")
	{
		defer os.Remove("/tmp/metrics.json")

		entity := Metrics{
			storage:                         localfs.NewPlaintextStorage("/tmp"),
			tenant:                          "1",
			promisedTransactions:            metrics.NewCounter(),
			promisedTransfers:               metrics.NewCounter(),
			committedTransactions:           metrics.NewCounter(),
			committedTransfers:              metrics.NewCounter(),
			rollbackedTransactions:          metrics.NewCounter(),
			rollbackedTransfers:             metrics.NewCounter(),
			transactionFinalizerCronLatency: metrics.NewTimer(),
		}

		require.Nil(t, entity.Persist())

		expected, err := entity.MarshalJSON()
		require.Nil(t, err)

		actual, err := ioutil.ReadFile("/tmp/metrics.1.json")
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
			storage:                         localfs.NewPlaintextStorage("/tmp"),
			tenant:                          "1",
			promisedTransactions:            metrics.NewCounter(),
			promisedTransfers:               metrics.NewCounter(),
			committedTransactions:           metrics.NewCounter(),
			committedTransfers:              metrics.NewCounter(),
			rollbackedTransactions:          metrics.NewCounter(),
			rollbackedTransfers:             metrics.NewCounter(),
			transactionFinalizerCronLatency: metrics.NewTimer(),
		}

		old.promisedTransactions.Inc(1)
		old.promisedTransfers.Inc(2)
		old.committedTransactions.Inc(3)
		old.committedTransfers.Inc(4)
		old.rollbackedTransactions.Inc(5)
		old.rollbackedTransfers.Inc(6)
		old.transactionFinalizerCronLatency.Update(time.Duration(7))

		data, err := old.MarshalJSON()
		require.Nil(t, err)

		require.Nil(t, ioutil.WriteFile("/tmp/metrics.1.json", data, 0444))

		entity := Metrics{
			storage:                         localfs.NewPlaintextStorage("/tmp"),
			tenant:                          "1",
			promisedTransactions:            metrics.NewCounter(),
			promisedTransfers:               metrics.NewCounter(),
			committedTransactions:           metrics.NewCounter(),
			committedTransfers:              metrics.NewCounter(),
			rollbackedTransactions:          metrics.NewCounter(),
			rollbackedTransfers:             metrics.NewCounter(),
			transactionFinalizerCronLatency: metrics.NewTimer(),
		}

		require.Nil(t, entity.Hydrate())

		assert.Equal(t, int64(1), entity.promisedTransactions.Count())
		assert.Equal(t, int64(2), entity.promisedTransfers.Count())
		assert.Equal(t, int64(3), entity.committedTransactions.Count())
		assert.Equal(t, int64(4), entity.committedTransfers.Count())
		assert.Equal(t, int64(5), entity.rollbackedTransactions.Count())
		assert.Equal(t, int64(6), entity.rollbackedTransfers.Count())
		assert.Equal(t, float64(7), entity.transactionFinalizerCronLatency.Percentile(0.95))
	}
}
