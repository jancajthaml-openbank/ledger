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

package actor

import (
	"github.com/jancajthaml-openbank/ledger-unit/model"

	system "github.com/jancajthaml-openbank/actor-system"
)

// TransactionStage enum INITIAL | PROMISE | COMMIT | ROLLBACK
type TransactionStage int

const (
	// INITIAL stage enum
	INITIAL TransactionStage = iota
	// PROMISE stage enum
	PROMISE
	// COMMIT stage enum
	COMMIT
	// ROLLBACK stage enum
	ROLLBACK
)

// TransactionState represent negotiation state of transaction actor
type TransactionState struct {
	Transaction     model.Transaction
	Negotiation     map[model.Account]string
	WaitFor         map[model.Account]bool
	OkResponses     int
	FailedResponses int
	Stage           TransactionStage
	ReplyTo         system.Coordinates
}

// NewTransactionState returns initial negotiation transaction actor state
func NewTransactionState() TransactionState {
	return TransactionState{
		OkResponses:     0,
		FailedResponses: 0,
		Stage:           INITIAL,
	}
}

func (state *TransactionState) markCommit(value interface{}) *model.Account {
	if state == nil {
		return nil
	}

	switch msg := value.(type) {

	case CommitWasAccepted:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.OkResponses++
		}
		return nil

	case CommitWasRejected:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}
		return nil

	case FatalErrored:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}
		return nil

	default:
		return nil

	}
}

func (state *TransactionState) markPromise(value interface{}) *model.Account {
	if state == nil {
		return nil
	}

	switch msg := value.(type) {

	case PromiseWasAccepted:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.OkResponses++
		}
		return nil

	case PromiseWasRejected:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}
		return nil

	case PromiseWasBounced:
		if _, exists := state.WaitFor[msg.Account]; exists {
			return &msg.Account
		}
		return nil

	case FatalErrored:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}
		return nil

	default:
		return nil

	}
}

func (state *TransactionState) markRollback(value interface{}) *model.Account {
	if state == nil {
		return nil
	}

	switch msg := value.(type) {

	case RollbackWasAccepted:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.OkResponses++
		}
		return nil

	case RollbackWasRejected:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}
		return nil

	case FatalErrored:
		if _, exists := state.WaitFor[msg.Account]; exists {
			delete(state.WaitFor, msg.Account)
			state.FailedResponses++
		}
		return nil

	default:
		return nil

	}
}

// Mark update negotiation state based on value
func (state *TransactionState) Mark(value interface{}) *model.Account {
	if state == nil {
		return nil
	}

	switch state.Stage {
	case INITIAL:
		return nil
	case PROMISE:
		return state.markPromise(value)
	case COMMIT:
		return state.markCommit(value)
	case ROLLBACK:
		return state.markRollback(value)
	default:
		return nil
	}
}

// ChangeStage zeroes out negotiation state and changes current stage
func (state *TransactionState) ChangeStage(nextStage TransactionStage) {
	if state == nil || state.Stage == nextStage {
		return
	}
	state.WaitFor = make(map[model.Account]bool)
	for account := range state.Negotiation {
		state.WaitFor[account] = true
	}
	state.Stage = nextStage
	state.OkResponses = 0
	state.FailedResponses = 0
}

// IsNegotiationFinished tells whenever negotiation is finished
func (state TransactionState) IsNegotiationFinished() bool {
	return len(state.Negotiation) <= (state.OkResponses + state.FailedResponses)
}

// PrepareNewForTransaction prepares state for new negotiation
func (state *TransactionState) PrepareNewForTransaction(transaction model.Transaction, requestedBy system.Coordinates) {
	if state == nil {
		return
	}
	negotiation := transaction.PrepareRemoteNegotiation()
	state.Transaction = transaction
	state.Transaction.State = model.StatusNew
	state.Negotiation = negotiation
	state.ChangeStage(PROMISE)
	state.ReplyTo = requestedBy
}
