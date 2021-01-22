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

import (
	"bytes"
	"reflect"
	"strings"
	"unsafe"
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
	var buffer bytes.Buffer

	buffer.WriteString(entity.State)
	buffer.WriteString("\n")

	for _, transfer := range entity.Transfers {
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
		buffer.WriteString("\n")
	}

	return buffer.Bytes()
}

// Deserialize transaction from persistent data
func (entity *Transaction) Deserialize(data []byte) {
	if entity == nil {
		return
	}

	// TODO improve performance of this function

	entity.Transfers = make([]Transfer, 0)

	// FIXME improve perf
	var j = bytes.IndexByte(data, '\n')

	entity.State = string(data[0:j])

	var i = j + 1
	var transfer []string

scan:
// FIXME improve perf
	j = bytes.IndexByte(data[i:], '\n')
	if j < 0 {
		if len(data) > 0 {
			// FIXME improve perf
			transfer = strings.SplitN(string(data[i:]), " ", 8)
			goto parse
		}
		return
	}
	j += i
	// FIXME improve perf
	transfer = strings.SplitN(string(data[i:j]), " ", 8)

parse:
	if len(transfer) != 8 {
		return
	}
	amount := new(Dec)
	amount.SetString(transfer[6])

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

	i = j + 1
	goto scan
}

// DeserializeState saves first line of data into Transaction State
func (entity *Transaction) DeserializeState(data []byte) {
	if entity == nil {
		return
	}
	var j = 0
	header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	for j < header.Len {
		if data[j] == '\n' {
			entity.State = *(*string)(unsafe.Pointer(&reflect.StringHeader{
				Data: header.Data,
				Len:  j,
			}))
			return
		}
		j++
	}
	return
}
