package model

import (
	"fmt"

	"testing"
)

func Benchmark_Serialize(b *testing.B) {
	entity := new(Transaction)
	entity.IDTransaction = "trn"
	entity.State = "committed"
	entity.Transfers = make([]Transfer, 10)

	for i := 0 ; i < len(entity.Transfers) ; i++ {
		trn := Transfer{
			IDTransfer: fmt.Sprintf("trx%d", i),
			Credit: Account{
				Tenant: "tenant",
				Name: "credit",
			},
			Debit: Account{
				Tenant: "tenant",
				Name: "debit",
			},
			ValueDate: "2018-03-04T17:08:22Z",
			Amount: new(Dec),
			Currency: "CUR",
		}
		
		entity.Transfers[i] = trn
	}

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entity.Serialize()
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
