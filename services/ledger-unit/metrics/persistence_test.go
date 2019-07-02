package metrics

import (
	"io/ioutil"
	"os"
	"testing"

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
		assert.EqualError(t, entity.Persist(), "json: error calling MarshalJSON for type *metrics.Metrics: cannot marshall nil references")
	}

	t.Log("error when cannot open tempfile for writing")
	{
		entity := Metrics{
			output:                 "/sys/kernel/security",
			promisedTransactions:   metrics.NewCounter(),
			promisedTransfers:      metrics.NewCounter(),
			committedTransactions:  metrics.NewCounter(),
			committedTransfers:     metrics.NewCounter(),
			rollbackedTransactions: metrics.NewCounter(),
			rollbackedTransfers:    metrics.NewCounter(),
			forwardedTransactions:  metrics.NewCounter(),
			forwardedTransfers:     metrics.NewCounter(),
		}

		assert.NotNil(t, entity.Persist())
	}

	t.Log("happy path")
	{
		tmpfile, err := ioutil.TempFile(os.TempDir(), "test_metrics_persist")

		require.Nil(t, err)
		defer os.Remove(tmpfile.Name())

		entity := Metrics{
			output:                 tmpfile.Name(),
			promisedTransactions:   metrics.NewCounter(),
			promisedTransfers:      metrics.NewCounter(),
			committedTransactions:  metrics.NewCounter(),
			committedTransfers:     metrics.NewCounter(),
			rollbackedTransactions: metrics.NewCounter(),
			rollbackedTransfers:    metrics.NewCounter(),
			forwardedTransactions:  metrics.NewCounter(),
			forwardedTransfers:     metrics.NewCounter(),
		}

		require.Nil(t, entity.Persist())

		expected, err := entity.MarshalJSON()
		require.Nil(t, err)

		actual, err := ioutil.ReadFile(tmpfile.Name())
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
		tmpfile, err := ioutil.TempFile(os.TempDir(), "test_metrics_hydrate")

		require.Nil(t, err)
		defer os.Remove(tmpfile.Name())

		old := Metrics{
			promisedTransactions:   metrics.NewCounter(),
			promisedTransfers:      metrics.NewCounter(),
			committedTransactions:  metrics.NewCounter(),
			committedTransfers:     metrics.NewCounter(),
			rollbackedTransactions: metrics.NewCounter(),
			rollbackedTransfers:    metrics.NewCounter(),
			forwardedTransactions:  metrics.NewCounter(),
			forwardedTransfers:     metrics.NewCounter(),
		}

		old.promisedTransactions.Inc(1)
		old.promisedTransfers.Inc(2)
		old.committedTransactions.Inc(3)
		old.committedTransfers.Inc(4)
		old.rollbackedTransactions.Inc(5)
		old.rollbackedTransfers.Inc(6)
		old.forwardedTransactions.Inc(7)
		old.forwardedTransfers.Inc(8)

		data, err := old.MarshalJSON()
		require.Nil(t, err)

		require.Nil(t, ioutil.WriteFile(tmpfile.Name(), data, 0444))

		entity := Metrics{
			output:                 tmpfile.Name(),
			promisedTransactions:   metrics.NewCounter(),
			promisedTransfers:      metrics.NewCounter(),
			committedTransactions:  metrics.NewCounter(),
			committedTransfers:     metrics.NewCounter(),
			rollbackedTransactions: metrics.NewCounter(),
			rollbackedTransfers:    metrics.NewCounter(),
			forwardedTransactions:  metrics.NewCounter(),
			forwardedTransfers:     metrics.NewCounter(),
		}

		require.Nil(t, entity.Hydrate())

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
