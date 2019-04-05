// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	money "gopkg.in/inf.v0"
)

type negotiatonChunk struct {
	Currency string
	Key      Account
}

// IsSameAs represents equality check of two Transactions
func (entity *Transaction) IsSameAs(obj *Transaction) bool {
	if entity == nil || obj == nil {
		return false
	}

	if entity.IDTransaction != obj.IDTransaction {
		return false
	}

	leftLen := len(entity.Transfers)
	rightLen := len(obj.Transfers)

	if leftLen != rightLen {
		return false
	}

	x := make([]string, leftLen)
	y := make([]string, rightLen)

	for i, e := range entity.Transfers {
		x[i] = e.Credit.Tenant + "/" + e.Credit.Name + "/" + e.Debit.Tenant + "/" + e.Debit.Name + "/" + e.Amount.String() + "/" + e.Currency
	}

	for i, e := range obj.Transfers {
		y[i] = e.Credit.Tenant + "/" + e.Credit.Name + "/" + e.Debit.Tenant + "/" + e.Debit.Name + "/" + e.Amount.String() + "/" + e.Currency
	}

	visited := make([]bool, len(y))
	for i := 0; i < len(x); i++ {
		found := false
		for j := 0; j < len(y); j++ {
			if visited[j] {
				continue
			}
			if y[j] == x[i] {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// PrepareRemoteNegotiation prepares negotiation of promises for all related accounts
func (entity *Transaction) PrepareRemoteNegotiation() map[Account]string {
	if entity == nil {
		return nil
	}

	var result = make(map[Account]string)

	chunks := make(map[negotiatonChunk][]*money.Dec)

	for _, transfer := range entity.Transfers {
		keyDebit := negotiatonChunk{
			Currency: transfer.Currency,
			Key:      transfer.Debit,
		}

		keyCredit := negotiatonChunk{
			Currency: transfer.Currency,
			Key:      transfer.Credit,
		}

		chunks[keyCredit] = append(chunks[keyCredit], transfer.Amount)
		chunks[keyDebit] = append(chunks[keyDebit], new(money.Dec).Neg(transfer.Amount))
	}

	for chunk, amounts := range chunks {
		var acc *money.Dec = new(money.Dec)

		for _, amount := range amounts {
			acc = new(money.Dec).Add(acc, amount)
		}

		result[chunk.Key] = entity.IDTransaction + " " + acc.String() + " " + chunk.Currency
	}

	return result
}
