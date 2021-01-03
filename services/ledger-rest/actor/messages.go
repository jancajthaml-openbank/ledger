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

package actor

import (
	"github.com/jancajthaml-openbank/ledger-rest/model"
	"strings"
	"time"
)

const (
	// ReqCreateTransaction ledger message request code for "Create Transaction"
	ReqCreateTransaction = "NT"
	// RespCreateTransaction ledger message response code for "Transaction Committed"
	RespCreateTransaction = "T0"
	// RespTransactionRace ledger message response code for "Transaction Race"
	RespTransactionRace = "T1"
	// RespTransactionRefused ledger message response code for "Transaction Refused"
	RespTransactionRefused = "T2"
	// RespTransactionRejected ledger message response code for "Transaction Rollbacked"
	RespTransactionRejected = "T3"
	// RespTransactionDuplicate ledger message response code for "Transaction Duplicate"
	RespTransactionDuplicate = "T4"
	// RespTransactionMissing ledger message response code for "Transaction Missing"
	RespTransactionMissing = "T5"
	// FatalError ledger message response code for "Error"
	FatalError = "EE"
)

// CreateTransactionMessage is message for creation of new transaction
func CreateTransactionMessage(transaction model.Transaction) string {
	var buffer strings.Builder

	numOfTransfers := len(transaction.Transfers)

	for idx, transfer := range transaction.Transfers {
		buffer.WriteString(transfer.IDTransfer)
		buffer.WriteString(";")
		buffer.WriteString(transfer.Credit.Tenant)
		buffer.WriteString(";")
		buffer.WriteString(transfer.Credit.Name)
		buffer.WriteString(";")
		buffer.WriteString(transfer.Debit.Tenant)
		buffer.WriteString(";")
		buffer.WriteString(transfer.Debit.Name)
		buffer.WriteString(";")
		buffer.WriteString(transfer.Amount)
		buffer.WriteString(";")
		buffer.WriteString(transfer.Currency)
		buffer.WriteString(";")
		buffer.WriteString(transfer.ValueDate.Format(time.RFC3339))
		if idx != numOfTransfers-1 {
			buffer.WriteString(" ")
		}
	}

	return ReqCreateTransaction + " " + transaction.IDTransaction + " " + buffer.String()
}
