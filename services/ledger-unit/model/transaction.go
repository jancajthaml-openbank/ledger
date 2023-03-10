// Copyright (c) 2016-2023, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"bytes"

	"github.com/jancajthaml-openbank/ledger-unit/support/cast"
)

const (
	// StatusNew represents NEW transaction
	StatusNew = "new"
	// StatusAccepted represents ACCEPTED transaction
	StatusAccepted = "accepted"
	// StatusRejected represents REJECTED transaction
	StatusRejected = "rejected"
	// StatusCommitted represents COMMITTED transaction
	StatusCommitted = "committed"
	// StatusRollbacked represents ROLLBACKED transaction
	StatusRollbacked = "rollbacked"
)

// Transfer represents ingress/egress message of transfer
type Transfer struct {
	IDTransfer string
	Credit     Account
	Debit      Account
	ValueDate  string
	Amount     *Dec
	Currency   string
}

// Transaction represents egress message of transaction
type Transaction struct {
	IDTransaction string
	State         string
	Transfers     []Transfer
}

// Serialize transaction to binary data
func (entity *Transaction) Serialize() []byte {
	if entity == nil {
		return nil
	}
	var buffer bytes.Buffer

	buffer.WriteString(entity.State)

	for _, transfer := range entity.Transfers {
		buffer.WriteString("\n")
		buffer.WriteString(transfer.IDTransfer)
		buffer.WriteString(" ")
		buffer.WriteString(transfer.Credit.Tenant)
		buffer.WriteString(" ")
		buffer.WriteString(transfer.Credit.Name)
		buffer.WriteString(" ")
		buffer.WriteString(transfer.Debit.Tenant)
		buffer.WriteString(" ")
		buffer.WriteString(transfer.Debit.Name)
		buffer.WriteString(" ")
		buffer.WriteString(transfer.ValueDate)
		buffer.WriteString(" ")
		buffer.WriteString(transfer.Amount.String())
		buffer.WriteString(" ")
		buffer.WriteString(transfer.Currency)
	}

	return buffer.Bytes()
}

// Deserialize transaction from persistent data
func (entity *Transaction) Deserialize(data []byte) {
	if entity == nil {
		return
	}

	var (
		i int
		j int
		k int
		l = len(data)
	)

	entity.Transfers = make([]Transfer, 0)

	for ; j < l && data[j] != '\n'; j++ {
	}

	entity.State = cast.BytesToString(data[0:j])

	j++
	if j >= l {
		return
	}
	i = j
	transfer := make([]string, 8)

scan:
	if i >= l {
		return
	}
	idx := 0
	k = i
	for ; k < l && idx < 8; k++ {
		if data[k] == ' ' || data[k] == '\n' {
			transfer[idx] = cast.BytesToString(data[i:k])
			idx++
			i = k + 1
		}
	}
	if k == l && idx < 8 {
		transfer[idx] = cast.BytesToString(data[i:])
		i = k + 1
	}

	amount := new(Dec)
	if !amount.SetString(transfer[6]) {
		return
	}

	entity.Transfers = append(entity.Transfers, Transfer{
		IDTransfer: transfer[0],
		Credit: Account{
			Tenant: transfer[1],
			Name:   transfer[2],
		},
		Debit: Account{
			Tenant: transfer[3],
			Name:   transfer[4],
		},
		ValueDate: transfer[5],
		Amount:    amount,
		Currency:  transfer[7],
	})

	goto scan
}

// DeserializeState saves first line of data into Transaction State
func (entity *Transaction) DeserializeState(data []byte) {
	if entity == nil {
		return
	}
	var j = 0
	var l = len(data)
	for ; j < l && data[j] != '\n'; j++ {
	}
	entity.State = cast.BytesToString(data[0:j])
	return
}
