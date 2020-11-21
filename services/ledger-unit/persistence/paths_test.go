package persistence

import (
  "fmt"
  "testing"

  "github.com/stretchr/testify/assert"
)

func TestTransactionsPath(t *testing.T) {
  path := TransactionsPath()

  assert.Equal(t, "transaction", path)
}

func TestTransactionPath(t *testing.T) {
  transaction := "trn_1"

  path := TransactionPath(transaction)
  expected := fmt.Sprintf("transaction/%s", transaction)

  assert.Equal(t, expected, path)
}

func BenchmarkTransactionsPath(b *testing.B) {
  b.ResetTimer()
  b.ReportAllocs()
  for n := 0; n < b.N; n++ {
    TransactionsPath()
  }
}

func BenchmarkTransactionPath(b *testing.B) {
  b.ResetTimer()
  b.ReportAllocs()
  for n := 0; n < b.N; n++ {
    TransactionPath("X")
  }
}
