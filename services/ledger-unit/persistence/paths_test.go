package persistence

import (
  "fmt"
  "testing"

  "github.com/stretchr/testify/assert"
)

func TestRootPath(t *testing.T) {
  path := RootPath()

  assert.Equal(t, "transaction", path)
}

func TestTransactionPath(t *testing.T) {
  transaction := "trn_1"

  path := TransactionPath(transaction)
  expected := fmt.Sprintf("transaction/%s", transaction)

  assert.Equal(t, expected, path)
}

func BenchmarkRootPath(b *testing.B) {
  b.ResetTimer()
  b.ReportAllocs()
  for n := 0; n < b.N; n++ {
    RootPath()
  }
}

func BenchmarkTransactionPath(b *testing.B) {
  b.ResetTimer()
  b.ReportAllocs()
  for n := 0; n < b.N; n++ {
    TransactionPath("X")
  }
}
