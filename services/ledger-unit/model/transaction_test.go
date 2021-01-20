package model

import "testing"

func Benchmark_DeserializeState(b *testing.B) {
	entity := new(Transaction)
	data := []byte("committed\n")

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		entity.DeserializeState(data)
	}
}
