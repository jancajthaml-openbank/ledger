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

import (
	"reflect"

	"github.com/jancajthaml-openbank/ledger-unit/model"
	"github.com/jancajthaml-openbank/ledger-unit/persistence"

	system "github.com/jancajthaml-openbank/actor-system"
)

func InitialTransaction(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(TransactionState)

		switch msg := context.Data.(type) {
		case model.Transaction:
			if state.Ready {
				s.SendMessage(
					RespTransactionRace+" "+msg.IDTransaction,
					state.ReplyTo,
					context.Receiver,
				)
				log.Warnf("Transaction already in progress %s", msg.IDTransaction)
				return
			}
			state.Prepare(msg, context.Sender)

		default:
			s.SendMessage(
				FatalError,
				state.ReplyTo,
				context.Receiver,
			)
			log.Warnf("Invalid message in InitialTransaction")
			return
		}

		if persistence.PersistTransaction(s.Storage, &state.Transaction) == nil {
			current, err := persistence.LoadTransaction(s.Storage, state.Transaction.IDTransaction)
			if err != nil {
				s.SendMessage(
					FatalError,
					state.ReplyTo,
					context.Receiver,
				)
				return
			}

			switch current.State {

			case persistence.StatusCommitted, persistence.StatusRollbacked:

				if state.Transaction.IsSameAs(current) {
					if current.State == persistence.StatusCommitted {
						s.SendMessage(
							RespCreateTransaction+" "+state.Transaction.IDTransaction,
							state.ReplyTo,
							context.Receiver,
						)
					} else {
						s.SendMessage(
							RespTransactionRejected+" "+state.Transaction.IDTransaction+" "+state.Transaction.State,
							state.ReplyTo,
							context.Receiver,
						)
					}
				} else {
					s.SendMessage(
						RespTransactionDuplicate+" "+state.Transaction.IDTransaction,
						state.ReplyTo,
						context.Receiver,
					)
				}

			default:
				s.SendMessage(
					RespTransactionRace+" "+state.Transaction.IDTransaction,
					state.ReplyTo,
					context.Receiver,
				)

			}

			return
		}

		s.Metrics.TransactionPromised(len(state.Transaction.Transfers))

		for account, task := range state.Negotiation {
			s.SendMessage(
				PromiseOrder+" "+task,
				system.Coordinates{
					Region: "VaultUnit/" + account.Tenant,
					Name:   account.Name,
				},
				context.Receiver,
			)
		}

		state.ResetMarks()
		context.Self.Become(state, PromisingTransaction(s))
		log.Infof("~ %v Start->Promise", state.Transaction.IDTransaction)
	}
}

func PromisingTransaction(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(TransactionState)

		switch msg := context.Data.(type) {

		case PromiseWasAccepted:
			log.Debugf("~ %v Promise Accepted %s", state.Transaction.IDTransaction, msg.Account)

		case PromiseWasRejected:
			log.Debugf("~ %v Promise Rejected %s %s", state.Transaction.IDTransaction, msg.Account, msg.Reason)

		case FatalErrored:
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
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)
			log.Debugf("~ %v Promise Rejected All", state.Transaction.IDTransaction)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		if state.FailedResponses > 0 {
			log.Debugf("~ %v Promise Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)

			state.Transaction.State = persistence.StatusRejected

			if persistence.UpdateTransaction(s.Storage, &state.Transaction) == nil {
				s.SendMessage(
					RespTransactionRefused+" "+state.Transaction.IDTransaction,
					state.ReplyTo,
					context.Receiver,
				)
				log.Warnf("~ %v Promise failed to reject transaction", state.Transaction.IDTransaction)
				s.UnregisterActor(context.Sender.Name)
				return
			}

			log.Infof("~ %v Promise->Rollback", state.Transaction.IDTransaction)

			state.ResetMarks()
			context.Self.Become(state, RollbackingTransaction(s))

			for account, task := range state.Negotiation {
				s.SendMessage(
					RollbackOrder+" "+task,
					system.Coordinates{
						Region: "VaultUnit/" + account.Tenant,
						Name:   account.Name,
					},
					context.Receiver,
				)
			}

			return
		}

		log.Debugf("~ %v Promise Accepted All", state.Transaction.IDTransaction)

		state.Transaction.State = persistence.StatusAccepted

		// FIXME possible null here
		if persistence.UpdateTransaction(s.Storage, &state.Transaction) == nil {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)
			log.Warnf("~ %v Promise failed to accept transaction", state.Transaction.IDTransaction)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		for account, task := range state.Negotiation {
			s.SendMessage(
				CommitOrder+" "+task,
				system.Coordinates{
					Region: "VaultUnit/" + account.Tenant,
					Name:   account.Name,
				},
				context.Receiver,
			)
		}

		state.ResetMarks()
		context.Self.Become(state, CommitingTransaction(s))
		log.Infof("~ %v Promise->Commit", state.Transaction.IDTransaction)
		return
	}
}

func CommitingTransaction(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(TransactionState)

		switch msg := context.Data.(type) {

		case CommitWasAccepted:
			log.Debugf("~ %v Commit Accepted %s", state.Transaction.IDTransaction, msg.Account)

		case CommitWasRejected:
			log.Debugf("~ %v Commit Rejected %s %s", state.Transaction.IDTransaction, msg.Account, msg.Reason)

		case FatalErrored:
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

			state.Transaction.State = persistence.StatusRejected

			if persistence.UpdateTransaction(s.Storage, &state.Transaction) == nil {
				s.SendMessage(
					RespTransactionRefused+" "+state.Transaction.IDTransaction,
					state.ReplyTo,
					context.Receiver,
				)
				log.Warnf("~ %v Commit failed to reject transaction", state.Transaction.IDTransaction)
				s.UnregisterActor(context.Sender.Name)
				return
			}

			for account, task := range state.Negotiation {
				s.SendMessage(
					RollbackOrder+" "+task,
					system.Coordinates{
						Region: "VaultUnit/" + account.Tenant,
						Name:   account.Name,
					},
					context.Receiver,
				)
			}

			state.ResetMarks()
			context.Self.Become(state, RollbackingTransaction(s))
			log.Infof("~ %v Commit->Rollback", state.Transaction.IDTransaction)
			return
		}

		log.Debugf("~ %v Commit Accepted All", state.Transaction.IDTransaction)

		state.Transaction.State = persistence.StatusCommitted

		if persistence.UpdateTransaction(s.Storage, &state.Transaction) == nil {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)
			log.Warnf("~ %v Commit failed to commit transaction", state.Transaction.IDTransaction)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		var transfers []string
		for _, transfer := range state.Transaction.Transfers {
			transfers = append(transfers, transfer.IDTransfer)
		}

		s.Metrics.TransactionCommitted(len(state.Transaction.Transfers))
		s.SendMessage(
			RespCreateTransaction+" "+state.Transaction.IDTransaction,
			state.ReplyTo,
			context.Receiver,
		)
		log.Infof("~ %v Commit->End", state.Transaction.IDTransaction)
		s.UnregisterActor(context.Sender.Name)
		return
	}
}

func RollbackingTransaction(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(TransactionState)

		switch msg := context.Data.(type) {

		case RollbackWasAccepted:
			log.Debugf("~ %v Rollback Accepted %s", state.Transaction.IDTransaction, msg.Account)

		case RollbackWasRejected:
			log.Debugf("~ %v Rollback Rejected %s %s", state.Transaction.IDTransaction, msg.Account, msg.Reason)

		case FatalErrored:
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
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)
			log.Debugf("~ %v Rollback Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		log.Debugf("~ %v Rollback Accepted All", state.Transaction.IDTransaction)

		// FIXME
		//rollBackReason := "unknown"

		state.Transaction.State = persistence.StatusRollbacked

		if persistence.UpdateTransaction(s.Storage, &state.Transaction) == nil {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)
			log.Warnf("~ %v Rollback failed to rollback transaction", state.Transaction.IDTransaction)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		s.Metrics.TransactionRollbacked(len(state.Transaction.Transfers))

		s.SendMessage(
			RespTransactionRejected+" "+state.Transaction.IDTransaction+" "+state.Transaction.State,
			state.ReplyTo,
			context.Receiver,
		)
		log.Infof("~ %v Rollback->End", state.Transaction.IDTransaction)
		s.UnregisterActor(context.Sender.Name)
		return
	}
}
