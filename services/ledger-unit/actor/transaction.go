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
	"reflect"

	"github.com/jancajthaml-openbank/ledger-unit/daemon"
	"github.com/jancajthaml-openbank/ledger-unit/model"
	"github.com/jancajthaml-openbank/ledger-unit/persistence"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
)

func InitialTransaction(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.TransactionState)

		switch msg := context.Data.(type) {
		case model.Transaction:
			if state.Ready {
				log.Warnf("Transaction already in progress %s", msg.IDTransaction)
				s.SendRemote(context.Sender.Region, TransactionRaceMessage(context.Receiver.Name, context.Sender.Name, msg.IDTransaction))
				return
			}
			state.Prepare(msg)

		default:
			log.Warnf("Invalid message in InitialTransaction")
			s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			return
		}

		if persistence.PersistTransaction(s.Storage, &state.Transaction) == nil {
			trnState, stateReason := persistence.GetTransactionState(s.Storage, state.Transaction.IDTransaction)

			switch trnState {

			case model.StatusCommitted, model.StatusRollbacked:
				current := persistence.LoadTransaction(s.Storage, state.Transaction.IDTransaction)

				if state.Transaction.IsSameAs(current) {
					if trnState == model.StatusCommitted {
						s.SendRemote(context.Sender.Region, TransactionProcessedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
					} else {
						s.SendRemote(context.Sender.Region, TransactionRejectedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction, stateReason))
					}
				} else {
					s.SendRemote(context.Sender.Region, TransactionDuplicateMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
				}

			default:
				s.SendRemote(context.Sender.Region, TransactionRaceMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))

			}

			return
		}

		log.Debugf("~ %v Start->Promise", state.Transaction.IDTransaction)
		state.ResetMarks()
		context.Receiver.Become(state, PromisingTransaction(s))
		s.Metrics.TransactionPromised(len(state.Transaction.Transfers))

		for account, task := range state.Negotiation {
			s.SendRemote("VaultUnit/"+account.Tenant, PromiseOrderMessage(context.Receiver.Name, account.Name, task))
		}
	}
}

func PromisingTransaction(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.TransactionState)

		switch context.Data.(type) {
		case model.PromiseWasAccepted:
			log.Debugf("~ %v Promise Accepted", state.Transaction.IDTransaction)
			state.MarkOk()
		default:
			log.Debugf("~ %v Promise Rejected %+v", state.Transaction.IDTransaction, reflect.ValueOf(context.Data).Type())
			state.MarkFailed()
		}

		if !state.IsNegotiationFinished() {
			context.Receiver.Become(state, PromisingTransaction(s))
			return
		}

		if state.OkResponses == 0 {
			log.Debugf("~ %v Promise Rejected All", state.Transaction.IDTransaction)
			s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
			return
		} else if state.FailedResponses > 0 {
			log.Debugf("~ %v Promise Rejected Some %d of %d", state.Transaction.IDTransaction, state.FailedResponses, state.NegotiationLen)
			if !persistence.RejectTransaction(s.Storage, state.Transaction.IDTransaction) {
				log.Warnf("~ %v Promise failed to reject transaction", state.Transaction.IDTransaction)
				s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
				return
			}

			log.Debugf("~ %v Promise->Rollback", state.Transaction.IDTransaction)
			state.ResetMarks()
			context.Receiver.Become(state, RollbackingTransaction(s))

			for account, task := range state.Negotiation {
				s.SendRemote("VaultUnit/"+account.Tenant, RollbackOrderMessage(context.Receiver.Name, account.Name, task))
			}

			return
		}

		log.Debugf("~ %v Promise Accepted All", state.Transaction.IDTransaction)

		// FIXME possible null here
		if !persistence.AcceptTransaction(s.Storage, state.Transaction.IDTransaction) {
			log.Warnf("~ %v Promise failed to accept transaction", state.Transaction.IDTransaction)
			s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
			return
		}

		log.Debugf("~ %v Promise->Commit", state.Transaction.IDTransaction)
		state.ResetMarks()
		context.Receiver.Become(state, CommitingTransaction(s))

		for account, task := range state.Negotiation {
			s.SendRemote("VaultUnit/"+account.Tenant, CommitOrderMessage(context.Receiver.Name, account.Name, task))
		}
		return
	}
}

func CommitingTransaction(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.TransactionState)

		switch context.Data.(type) {
		case model.CommitWasAccepted:
			log.Debugf("~ %v Commit Accepted", state.Transaction.IDTransaction)
			state.MarkOk()
		default:
			log.Debugf("~ %v Commit Rejected %+v", state.Transaction.IDTransaction, reflect.ValueOf(context.Data).Type())
			state.MarkFailed()
		}

		if !state.IsNegotiationFinished() {
			context.Receiver.Become(state, CommitingTransaction(s))
			return
		}

		if state.FailedResponses > 0 {
			log.Debugf("~ %v Commit Rejected %d of %d", state.Transaction.IDTransaction, state.FailedResponses, state.NegotiationLen)
			if !persistence.RejectTransaction(s.Storage, state.Transaction.IDTransaction) {
				log.Warnf("~ %v Commit failed to reject transaction", state.Transaction.IDTransaction)
				s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
				return
			}

			log.Debugf("~ %v Commit->Rollback", state.Transaction.IDTransaction)
			state.ResetMarks()
			context.Receiver.Become(state, RollbackingTransaction(s))

			for account, task := range state.Negotiation {
				s.SendRemote("VaultUnit/"+account.Tenant, RollbackOrderMessage(context.Receiver.Name, account.Name, task))
			}
			return
		}

		log.Debugf("~ %v Commit Accepted All", state.Transaction.IDTransaction)

		if !persistence.CommitTransaction(s.Storage, state.Transaction.IDTransaction) {
			log.Warnf("~ %v Commit failed to commit transaction", state.Transaction.IDTransaction)
			s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
			s.UnregisterActor(context.Sender.Name)
			return
		}

		var transfers []string
		for _, transfer := range state.Transaction.Transfers {
			transfers = append(transfers, transfer.IDTransfer)
		}

		log.Debugf("~ %v Commit->End", state.Transaction.IDTransaction)
		s.SendRemote(context.Sender.Region, TransactionProcessedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))

		s.Metrics.TransactionCommitted(len(state.Transaction.Transfers))
		s.UnregisterActor(context.Sender.Name)
		return
	}
}

func RollbackingTransaction(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.TransactionState)

		switch context.Data.(type) {
		case model.RollbackWasAccepted:
			log.Debugf("~ %v Rollback Accepted", state.Transaction.IDTransaction)
			state.MarkOk()
		default:
			log.Debugf("~ %v Rollback Rejected %+v", state.Transaction.IDTransaction, reflect.ValueOf(context.Data).Type())
			state.MarkFailed()
		}

		if !state.IsNegotiationFinished() {
			context.Receiver.Become(state, RollbackingTransaction(s))
			return
		}

		if state.FailedResponses > 0 {
			log.Debugf("~ %v Rollback Rejected %d of %d", state.Transaction.IDTransaction, state.FailedResponses, state.NegotiationLen)
			s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
			s.UnregisterActor(context.Sender.Name)
			return
		}

		log.Debugf("~ %v Rollback Accepted All", state.Transaction.IDTransaction)

		rollBackReason := "unknown"

		if !persistence.RollbackTransaction(s.Storage, state.Transaction.IDTransaction, rollBackReason) {
			log.Warnf("~ %v Rollback failed to rollback transaction", state.Transaction.IDTransaction)
			s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
			s.UnregisterActor(context.Sender.Name)
			return
		}

		log.Debugf("~ %v Rollback->End", state.Transaction.IDTransaction)
		s.SendRemote(context.Sender.Region, TransactionRejectedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction, rollBackReason))

		s.Metrics.TransactionRollbacked(len(state.Transaction.Transfers))
		s.UnregisterActor(context.Sender.Name)
		return
	}
}
