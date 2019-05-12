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

package actor

import (
	"strings"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/model"
)

const (
	ReqCreateTransaction     = "NT"
	ReqForwardTransfer       = "FT"
	RespCreateTransaction    = "T0"
	RespTransactionRace      = "T1"
	RespTransactionRefused   = "T2"
	RespTransactionRejected  = "T3"
	RespTransactionDuplicate = "T4"
	RespTransactionMissing   = "T5"
	FatalError               = "EE"
)

func CreateTransactionMessage(tenant string, sender string, name string, transaction model.Transaction) string {
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
		buffer.WriteString(transfer.Amount.String())
		buffer.WriteString(";")
		buffer.WriteString(transfer.Currency)
		buffer.WriteString(";")
		buffer.WriteString(transfer.ValueDate.Format(time.RFC3339))
		if idx != numOfTransfers-1 {
			buffer.WriteString(" ")
		}
	}

	return "LedgerUnit/" + tenant + " LedgerRest " + name + " " + sender + " " + ReqCreateTransaction + " " + transaction.IDTransaction + " " + buffer.String()
}

func ForwardTransferMessage(tenant string, sender string, name string, transaction string, transfer string, forward model.TransferForward) string {
	return "LedgerUnit/" + tenant + " LedgerRest " + name + " " + sender + " " + ReqForwardTransfer + " " + transaction + " " + transfer + " " + forward.Side + " " + forward.TargetAccount.Tenant + ";" + forward.TargetAccount.Name
}
