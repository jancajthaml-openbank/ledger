package model

import (
	"fmt"

	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Serialize(t *testing.T) {

	t.Log("nil")
	{
		var entity *Transaction
		assert.Nil(t, entity.Serialize())
	}

	t.Log("without transfers")
	{
		entity := new(Transaction)
		entity.IDTransaction = "trn"
		entity.State = "committed"
		assert.Equal(t, "committed", string(entity.Serialize()))
	}

	t.Log("with transfers")
	{
		entity := new(Transaction)
		entity.IDTransaction = "trn"
		entity.State = "committed"
		entity.Transfers = make([]Transfer, 2)

		for i := 0; i < len(entity.Transfers); i++ {
			entity.Transfers[i] = Transfer{
				IDTransfer: fmt.Sprintf("x%d", i),
				Credit: Account{
					Tenant: "t",
					Name:   "c",
				},
				Debit: Account{
					Tenant: "t",
					Name:   "d",
				},
				ValueDate: "v",
				Amount:    new(Dec),
				Currency:  "CUR",
			}
		}

		assert.Equal(t, "committed\nx0 t c t d v 0.0 CUR\nx1 t c t d v 0.0 CUR", string(entity.Serialize()))
	}
}

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

		assert.Equal(t, "committed", entity.State)
	}

	t.Log("with transfers")
	{
		data := []byte("committed\nx0 t c t d v 0.0 CUR\nx1 t c t d v 0.0 CUR")

		entity := new(Transaction)
		entity.Deserialize(data)

		assert.Equal(t, "committed", entity.State)
		assert.Equal(t, 2, len(entity.Transfers))
	}
}

func Benchmark_Serialize(b *testing.B) {
	entity := new(Transaction)
	entity.IDTransaction = "trn"
	entity.State = "committed"
	entity.Transfers = make([]Transfer, 10)

	for i := 0; i < len(entity.Transfers); i++ {
		entity.Transfers[i] = Transfer{
			IDTransfer: fmt.Sprintf("trx%d", i),
			Credit: Account{
				Tenant: "tenant",
				Name:   "credit",
			},
			Debit: Account{
				Tenant: "tenant",
				Name:   "debit",
			},
			ValueDate: "2018-03-04T17:08:22Z",
			Amount:    new(Dec),
			Currency:  "CUR",
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entity.Serialize()
	}
}

func Benchmark_Deserialize(b *testing.B) {
	entity := new(Transaction)
	entity.IDTransaction = "trn"
	entity.State = "committed"
	entity.Transfers = make([]Transfer, 10)

	for i := 0; i < len(entity.Transfers); i++ {
		entity.Transfers[i] = Transfer{
			IDTransfer: fmt.Sprintf("trx%d", i),
			Credit: Account{
				Tenant: "tenant",
				Name:   "credit",
			},
			Debit: Account{
				Tenant: "tenant",
				Name:   "debit",
			},
			ValueDate: "2018-03-04T17:08:22Z",
			Amount:    new(Dec),
			Currency:  "CUR",
		}
	}

	data := entity.Serialize()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entity.Deserialize(data)
	}
}

func Benchmark_DeserializeState(b *testing.B) {
	entity := new(Transaction)
	data := []byte("committed\n")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entity.DeserializeState(data)
	}
}
