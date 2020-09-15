// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
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

	// PromiseOrder vault message request code for "Promise"
	PromiseOrder = "NP"
	// PromiseAccepted vault message response code for "Promise" accepted
	PromiseAccepted = "P1"
	// PromiseRejected vault message response code for "Promise" rejected
	PromiseRejected = "P2"

	// CommitOrder vault message request code for "Commit"
	CommitOrder = "NC"
	// CommitAccepted vault message response code for "Commit" accepted
	CommitAccepted = "C1"
	// CommitRejected vault message response code for "Commit" rejected
	CommitRejected = "C2"

	// RollbackOrder vault message request code for "Rollback"
	RollbackOrder = "NR"
	// RollbackAccepted vault message response code for "Rollback" accepted
	RollbackAccepted = "R1"
	// RollbackRejected vault message response code for "Rollback" rejected
	RollbackRejected = "R2"

	// FatalError vault message response code for "Error"
	FatalError = "EE"
)
