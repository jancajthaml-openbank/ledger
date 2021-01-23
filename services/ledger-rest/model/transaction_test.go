package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DeSerialize(t *testing.T) {

	t.Log("nil")
	{
		var entity *Transaction
		entity.Deserialize([]byte(""))
	}

	t.Log("without transfers")
	{
		data := []byte("committed")

		entity := new(Transaction)
		entity.Deserialize(data)

		assert.Equal(t, "committed", entity.Status)
	}

	t.Log("with transfers")
	{
		data := []byte("committed\nx0 t c t d 2018-03-04T17:08:22Z 0.0 CUR\nx1 t c t d 2018-03-04T17:08:22Z 0.0 CUR")

		entity := new(Transaction)
		entity.Deserialize(data)

		assert.Equal(t, "committed", entity.Status)
		assert.Equal(t, 2, len(entity.Transfers))
	}
}

func Benchmark_Deserialize(b *testing.B) {
	entity := new(Transaction)

	data := []byte("committed\nx0 t c t d 2018-03-04T17:08:22Z 0.0 CUR\nx1 t c t d 2018-03-04T17:08:22Z 0.0 CUR")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entity.Deserialize(data)
	}
}
