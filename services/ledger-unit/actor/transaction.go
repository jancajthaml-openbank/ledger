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
				s.SendRemote(TransactionRaceMessage(context, msg.IDTransaction))
				log.Warnf("Transaction already in progress %s", msg.IDTransaction)
				return
			}
			state.Prepare(msg)

		default:
			s.SendRemote(FatalErrorMessage(context))
			log.Warnf("Invalid message in InitialTransaction")
			return
		}

		if persistence.PersistTransaction(s.Storage, &state.Transaction) == nil {
			trnState, stateReason := persistence.GetTransactionState(s.Storage, state.Transaction.IDTransaction)

			switch trnState {

			case model.StatusCommitted, model.StatusRollbacked:
				current := persistence.LoadTransaction(s.Storage, state.Transaction.IDTransaction)

				if state.Transaction.IsSameAs(current) {
					if trnState == model.StatusCommitted {
						s.SendRemote(TransactionProcessedMessage(context, state.Transaction.IDTransaction))
					} else {
						s.SendRemote(TransactionRejectedMessage(context, state.Transaction.IDTransaction, stateReason))
					}
				} else {
					s.SendRemote(TransactionDuplicateMessage(context, state.Transaction.IDTransaction))
				}

			default:
				s.SendRemote(TransactionRaceMessage(context, state.Transaction.IDTransaction))

			}

			return
		}

		s.Metrics.TransactionPromised(len(state.Transaction.Transfers))

		for account, task := range state.Negotiation {
			s.SendRemote(PromiseOrderMessage(context, account, task))
		}

		state.ResetMarks()
		context.Self.Become(state, PromisingTransaction(s))
		log.Infof("~ %v Start->Promise", state.Transaction.IDTransaction)
	}
}

func PromisingTransaction(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.TransactionState)

		switch msg := context.Data.(type) {
		case model.PromiseWasAccepted:
			log.Debugf("~ %v Promise Accepted %s", state.Transaction.IDTransaction, msg.Account)
		case model.PromiseWasRejected:
			log.Debugf("~ %v Promise Rejected %s %s", state.Transaction.IDTransaction, msg.Account, msg.Reason)
		case model.FatalErrored:
			log.Debugf("~ %v Promise Errored %s", state.Transaction.IDTransaction, msg.Account)
		default:
			log.Debugf("~ %v Promise Invalid Message %+v / %+v", state.Transaction.IDTransaction, reflect.ValueOf(context.Data).Type(), context.Data)
		}

		state.Mark(context.Data)

		if !state.IsNegotiationFinished() {
			context.Self.Become(state, PromisingTransaction(s))
			return
		}

		if state.OkResponses == 0 {
			s.SendRemote(TransactionRefusedMessage(context, state.Transaction.IDTransaction))
			log.Debugf("~ %v Promise Rejected All", state.Transaction.IDTransaction)
			return
		}

		if state.FailedResponses > 0 {
			log.Debugf("~ %v Promise Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)
			if !persistence.RejectTransaction(s.Storage, state.Transaction.IDTransaction) {
				s.SendRemote(TransactionRefusedMessage(context, state.Transaction.IDTransaction))
				log.Warnf("~ %v Promise failed to reject transaction", state.Transaction.IDTransaction)
				return
			}

			log.Infof("~ %v Promise->Rollback", state.Transaction.IDTransaction)

			state.ResetMarks()
			context.Self.Become(state, RollbackingTransaction(s))

			for account, task := range state.Negotiation {
				s.SendRemote(RollbackOrderMessage(context, account, task))
			}

			return
		}

		log.Debugf("~ %v Promise Accepted All", state.Transaction.IDTransaction)

		// FIXME possible null here
		if !persistence.AcceptTransaction(s.Storage, state.Transaction.IDTransaction) {
			s.SendRemote(TransactionRefusedMessage(context, state.Transaction.IDTransaction))
			log.Warnf("~ %v Promise failed to accept transaction", state.Transaction.IDTransaction)
			return
		}

		for account, task := range state.Negotiation {
			s.SendRemote(CommitOrderMessage(context, account, task))
		}

		state.ResetMarks()
		context.Self.Become(state, CommitingTransaction(s))
		log.Infof("~ %v Promise->Commit", state.Transaction.IDTransaction)
		return
	}
}

func CommitingTransaction(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.TransactionState)

		switch msg := context.Data.(type) {
		case model.CommitWasAccepted:
			log.Debugf("~ %v Commit Accepted %s", state.Transaction.IDTransaction, msg.Account)
		case model.CommitWasRejected:
			log.Debugf("~ %v Commit Rejected %s %s", state.Transaction.IDTransaction, msg.Account, msg.Reason)
		case model.FatalErrored:
			log.Debugf("~ %v Commit Errored %s", state.Transaction.IDTransaction, msg.Account)
		default:
			log.Debugf("~ %v Commit Invalid Message %+v / %+v", state.Transaction.IDTransaction, reflect.ValueOf(context.Data).Type(), context.Data)
		}
		state.Mark(context.Data)

		if !state.IsNegotiationFinished() {
			context.Self.Become(state, CommitingTransaction(s))
			return
		}

		if state.FailedResponses > 0 {
			log.Debugf("~ %v Commit Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)
			if !persistence.RejectTransaction(s.Storage, state.Transaction.IDTransaction) {
				s.SendRemote(TransactionRefusedMessage(context, state.Transaction.IDTransaction))
				log.Warnf("~ %v Commit failed to reject transaction", state.Transaction.IDTransaction)
				return
			}

			for account, task := range state.Negotiation {
				s.SendRemote(RollbackOrderMessage(context, account, task))
			}

			state.ResetMarks()
			context.Self.Become(state, RollbackingTransaction(s))
			log.Infof("~ %v Commit->Rollback", state.Transaction.IDTransaction)
			return
		}

		log.Debugf("~ %v Commit Accepted All", state.Transaction.IDTransaction)

		if !persistence.CommitTransaction(s.Storage, state.Transaction.IDTransaction) {
			s.SendRemote(TransactionRefusedMessage(context, state.Transaction.IDTransaction))
			log.Warnf("~ %v Commit failed to commit transaction", state.Transaction.IDTransaction)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		var transfers []string
		for _, transfer := range state.Transaction.Transfers {
			transfers = append(transfers, transfer.IDTransfer)
		}

		s.Metrics.TransactionCommitted(len(state.Transaction.Transfers))

		s.SendRemote(TransactionProcessedMessage(context, state.Transaction.IDTransaction))
		log.Infof("~ %v Commit->End", state.Transaction.IDTransaction)
		s.UnregisterActor(context.Sender.Name)
		return
	}
}

func RollbackingTransaction(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.TransactionState)

		switch msg := context.Data.(type) {
		case model.RollbackWasAccepted:
			log.Debugf("~ %v Rollback Accepted %s", state.Transaction.IDTransaction, msg.Account)
		case model.RollbackWasRejected:
			log.Debugf("~ %v Rollback Rejected %s %s", state.Transaction.IDTransaction, msg.Account, msg.Reason)
		case model.FatalErrored:
			log.Debugf("~ %v Rollback Errored %s", state.Transaction.IDTransaction, msg.Account)
		default:
			log.Debugf("~ %v Rollback Invalid Message %+v / %+v", state.Transaction.IDTransaction, reflect.ValueOf(context.Data).Type(), context.Data)
		}
		state.Mark(context.Data)

		if !state.IsNegotiationFinished() {
			context.Self.Become(state, RollbackingTransaction(s))
			return
		}

		if state.FailedResponses > 0 {
			s.SendRemote(TransactionRefusedMessage(context, state.Transaction.IDTransaction))
			log.Debugf("~ %v Rollback Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		log.Debugf("~ %v Rollback Accepted All", state.Transaction.IDTransaction)

		rollBackReason := "unknown"

		if !persistence.RollbackTransaction(s.Storage, state.Transaction.IDTransaction, rollBackReason) {
			s.SendRemote(TransactionRefusedMessage(context, state.Transaction.IDTransaction))
			log.Warnf("~ %v Rollback failed to rollback transaction", state.Transaction.IDTransaction)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		s.Metrics.TransactionRollbacked(len(state.Transaction.Transfers))

		s.SendRemote(TransactionRejectedMessage(context, state.Transaction.IDTransaction, rollBackReason))
		log.Infof("~ %v Rollback->End", state.Transaction.IDTransaction)
		s.UnregisterActor(context.Sender.Name)
		return
	}
}
