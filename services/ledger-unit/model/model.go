// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"strings"

	money "gopkg.in/inf.v0"
)

type PromiseWasAccepted struct{}

type CommitWasAccepted struct{}

type RollbackWasAccepted struct{}

type TransactionState struct {
	Transaction     Transaction
	Negotiation     map[Account]string
	NegotiationLen  int
	OkResponses     int
	FailedResponses int
	Ready           bool
}

func NewTransactionState() TransactionState {
	return TransactionState{
		OkResponses:     0,
		FailedResponses: 0,
		Ready:           false,
	}
}

func (state *TransactionState) MarkOk() {
	if state == nil {
		return
	}
	state.OkResponses += 1
}

func (state *TransactionState) MarkFailed() {
	if state == nil {
		return
	}
	state.FailedResponses += 1
}

func (state *TransactionState) ResetMarks() {
	if state == nil {
		return
	}
	state.OkResponses = 0
	state.FailedResponses = 0
}

func (state TransactionState) IsNegotiationFinished() bool {
	return state.NegotiationLen <= (state.OkResponses + state.FailedResponses)
}

func (state *TransactionState) Prepare(transaction Transaction) {
	if state == nil {
		return
	}

	negotiation := transaction.PrepareRemoteNegotiation()
	state.Transaction = transaction
	state.Negotiation = negotiation
	state.OkResponses = 0
	state.FailedResponses = 0
	state.NegotiationLen = len(negotiation)
	state.Ready = true
}

type ForwardState struct {
	Forward TransferForward
	TransactionState
}

func NewForwardState() ForwardState {
	return ForwardState{
		TransactionState: NewTransactionState(),
	}
}

// Transfer represents ingress/egress message of transfer
type Transfer struct {
	IDTransfer string
	Credit     Account
	Debit      Account
	ValueDate  string
	Amount     *money.Dec
	Currency   string
}

type TransferForward struct {
	IDTransaction string
	IDTransfer    string
	Side          string
	Target        Account
}

type Account struct {
	Tenant string
	Name   string
}

// Transaction represents egress message of transaction
type Transaction struct {
	IDTransaction string
	Transfers     []Transfer
}

// Serialise transaction to binary data
func (entity *Transaction) Serialise() []byte {
	var buffer bytes.Buffer

	buffer.WriteString(entity.IDTransaction)
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
		buffer.WriteString(strings.TrimRight(strings.TrimRight(transfer.Amount.String(), "0"), "."))
		buffer.WriteString(" ")
		buffer.WriteString(transfer.Currency)
		buffer.WriteString("\n")
	}

	return buffer.Bytes()
}

// Deserialise transaction from binary data
func (entity *Transaction) Deserialise(data []byte) {
	lines := strings.Split(string(data), "\n")
	entity.IDTransaction = lines[0]
	entity.Transfers = make([]Transfer, len(lines)-2)

	for i := range entity.Transfers {
		transfer := strings.SplitN(lines[i+1], " ", 8)

		amount, _ := new(money.Dec).SetString(transfer[6])

		entity.Transfers[i] = Transfer{
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
		}
	}

	return
}
