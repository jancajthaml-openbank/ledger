// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
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

type negotiatonChunk struct {
	Currency string
	Key      Account
}

// IsSameAs represents equality check of two Transactions
func (original *Transaction) IsSameAs(candidate *Transaction) bool {
	if original == nil || candidate == nil {
		return false
	}

	if original.IDTransaction != candidate.IDTransaction {
		return false
	}

	originalLen := len(original.Transfers)
	candidateLen := len(candidate.Transfers)

	if originalLen != candidateLen {
		return false
	}

	confirmed := make(map[int]bool)
	var found bool

	for _, left := range candidate.Transfers {
		found = false
		for j, right := range original.Transfers {
			if _, ok := confirmed[j]; ok {
				continue
			}
			if left.IDTransfer != right.IDTransfer {
				continue
			}
			if left.Currency != right.Currency {
				continue
			}
			if left.Credit.Tenant != right.Credit.Tenant {
				continue
			}
			if left.Credit.Name != right.Credit.Name {
				continue
			}
			if left.Debit.Tenant != right.Debit.Tenant {
				continue
			}
			if left.Debit.Name != right.Debit.Name {
				continue
			}
			diff := new(Dec)
			diff.Add(left.Amount)
			diff.Sub(right.Amount)
			if diff.Sign() != 0 {
				continue
			}
			confirmed[j] = true
			found = true
		}
		if !found {
			return false
		}
	}

	return len(confirmed) == originalLen
}

// PrepareRemoteNegotiation prepares negotiation of promises for all related accounts
func (original *Transaction) PrepareRemoteNegotiation() map[Account]string {
	if original == nil {
		return nil
	}

	var result = make(map[Account]string)

	chunks := make(map[negotiatonChunk][]*Dec)

	for _, transfer := range original.Transfers {
		keyDebit := negotiatonChunk{
			Currency: transfer.Currency,
			Key:      transfer.Debit,
		}

		keyCredit := negotiatonChunk{
			Currency: transfer.Currency,
			Key:      transfer.Credit,
		}

		debitAmount := new(Dec)
		debitAmount.Sub(transfer.Amount)

		chunks[keyCredit] = append(chunks[keyCredit], transfer.Amount)
		chunks[keyDebit] = append(chunks[keyDebit], debitAmount)
	}

	for chunk, amounts := range chunks {
		acc := new(Dec)
		for _, amount := range amounts {
			acc.Add(amount)
		}
		result[chunk.Key] = original.IDTransaction + " " + acc.String() + " " + chunk.Currency
	}

	return result
}
