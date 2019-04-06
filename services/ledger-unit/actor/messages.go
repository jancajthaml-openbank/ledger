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
	"github.com/jancajthaml-openbank/ledger-unit/model"
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

	// PromiseOrder vault message request code for "Promise"
	PromiseOrder = model.EventPromise + "X"
	// CommitOrder vault message request code for "Commit"
	CommitOrder = model.EventCommit + "X"
	// RollbackOrder vault message request code for "Rollback"
	RollbackOrder = model.EventRollback + "X"
	// PromiseAccepted vault message response code for "Promise"
	PromiseAccepted = "X" + model.EventPromise
	// CommitAccepted vault message response code for "Commit"
	CommitAccepted = "X" + model.EventCommit
	// RollbackAccepted vault message response code for "Rollback"
	RollbackAccepted = "X" + model.EventRollback
	// FatalError vault message response code for "Error"
	FatalError = "EE"
)

// FatalErrorMessage is reply message carrying failure
func FatalErrorMessage(sender, name string) string {
	return name + " " + sender + " " + FatalError
}

// PromiseOrderMessage is message for negotiation of promise transaction
func PromiseOrderMessage(sender, name, msg string) string {
	return name + " " + sender + " " + PromiseOrder + " " + msg
}

// CommitOrderMessage is message for negotiation of commit transaction
func CommitOrderMessage(sender, name, msg string) string {
	return name + " " + sender + " " + CommitOrder + " " + msg
}

// RollbackOrderMessage is message for negotiation of rollback transaction
func RollbackOrderMessage(sender, name, msg string) string {
	return name + " " + sender + " " + RollbackOrder + " " + msg
}

func TransactionRejectedMessage(sender, name, id, reason string) string {
	return name + " " + sender + " " + RespTransactionRejected + " " + id + " " + reason
}

func TransactionProcessedMessage(sender, name, id string) string {
	return name + " " + sender + " " + RespCreateTransaction + " " + id
}

func TransactionRefusedMessage(sender, name, id string) string {
	return name + " " + sender + " " + RespTransactionRefused + " " + id
}

func TransactionDuplicateMessage(sender, name, id string) string {
	return name + " " + sender + " " + RespTransactionDuplicate + " " + id
}

func TransactionRaceMessage(sender, name, id string) string {
	return name + " " + sender + " " + RespTransactionRace + " " + id
}

func TransactionMissingMessage(sender, name, id string) string {
	return name + " " + sender + " " + RespTransactionMissing + " " + id
}
