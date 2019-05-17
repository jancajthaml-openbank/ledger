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

type FatalErrored struct {
	Account Account
}

type PromiseWasAccepted struct {
	Account Account
}

type PromiseWasRejected struct {
	Account Account
	Reason  string
}

type CommitWasAccepted struct {
	Account Account
}

type CommitWasRejected struct {
	Account Account
	Reason  string
}

type RollbackWasAccepted struct {
	Account Account
}

type RollbackWasRejected struct {
	Account Account
	Reason  string
}

type TransactionState struct {
	Transaction     Transaction
	Negotiation     map[Account]string
	WaitFor         map[Account]interface{}
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

func (state *TransactionState) Mark(response interface{}) {
	if state == nil {
		return
	}

	switch msg := response.(type) {

	case PromiseWasAccepted:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.OkResponses++
		}

	case CommitWasAccepted:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.OkResponses++
		}

	case RollbackWasAccepted:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.OkResponses++
		}

	case PromiseWasRejected:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}

	case CommitWasRejected:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}

	case RollbackWasRejected:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}

	case FatalErrored:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}

	}
}

func (state *TransactionState) ResetMarks() {
	if state == nil {
		return
	}

	state.WaitFor = make(map[Account]interface{})

	for account := range state.Negotiation {
		state.WaitFor[account] = nil
	}

	state.OkResponses = 0
	state.FailedResponses = 0

}

func (state TransactionState) IsNegotiationFinished() bool {
	return len(state.Negotiation) <= (state.OkResponses + state.FailedResponses)
}

func (state *TransactionState) Prepare(transaction Transaction) {
	if state == nil {
		return
	}

	negotiation := transaction.PrepareRemoteNegotiation()
	state.Transaction = transaction
	state.Negotiation = negotiation
	state.ResetMarks()
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

func (s Account) String() string {
	return s.Tenant + "/" + s.Name
}

// Transaction represents egress message of transaction
type Transaction struct {
	IDTransaction string
	State         string
	Transfers     []Transfer
}

// Serialise transaction to binary data
func (entity *Transaction) Serialise() []byte {
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
	entity.State = lines[0]
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
