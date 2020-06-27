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
				log.WithField("transaction", state.Transaction.IDTransaction).Warn("Already in progress")
				return
			}
			state.PrepareNewForTransaction(msg, context.Sender)

		default:
			s.SendMessage(FatalError, state.ReplyTo, context.Receiver)
			return
		}

		err := persistence.CreateTransaction(s.Storage, &state.Transaction)
		if err != nil {
			current, err := persistence.LoadTransaction(s.Storage, state.Transaction.IDTransaction)
			if err != nil {
				s.SendMessage(FatalError, state.ReplyTo, context.Receiver)
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

		log.WithField("transaction", state.Transaction.IDTransaction).Debug("Start->Promise")
	}
}

func PromisingTransaction(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(TransactionState)
		state.Mark(context.Data)
		if !state.IsNegotiationFinished() {
			context.Self.Become(state, PromisingTransaction(s))
			return
		}

		if state.FailedResponses > 0 {
			state.Transaction.State = persistence.StatusRejected
			err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
			if err != nil {
				log.WithField("transaction", state.Transaction.IDTransaction).Errorf("Promise failed to update transaction %+v", err)
				s.SendMessage(
					RespTransactionRefused+" "+state.Transaction.IDTransaction,
					state.ReplyTo,
					context.Receiver,
				)
				s.UnregisterActor(context.Sender.Name)
				return
			}
		}

		if state.OkResponses == 0 {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)
			log.WithField("transaction", state.Transaction.IDTransaction).Debug("Promise Rejected All")
			return
		}

		if state.FailedResponses > 0 {
			log.WithField("transaction", state.Transaction.IDTransaction).Debugf("Promise Rejected Some [total: %d, accepted: %d, rejected: %d]", len(state.Negotiation), state.FailedResponses, state.OkResponses)
			log.WithField("transaction", state.Transaction.IDTransaction).Debug("Promise->Rollback")

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

		log.WithField("transaction", state.Transaction.IDTransaction).Debug("Promise Accepted All")

		state.Transaction.State = persistence.StatusAccepted

		err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
		if err != nil {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)

			log.WithField("transaction", state.Transaction.IDTransaction).Warn("Promise failed to accept transaction")

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
		log.WithField("transaction", state.Transaction.IDTransaction).Debug("Promise->Commit")
		return
	}
}

func CommitingTransaction(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(TransactionState)
		state.Mark(context.Data)
		if !state.IsNegotiationFinished() {
			context.Self.Become(state, CommitingTransaction(s))
			return
		}

		if state.FailedResponses > 0 {
			log.WithField("transaction", state.Transaction.IDTransaction).Debugf("Commit Rejected Some [total: %d, accepted: %d, rejected: %d]", len(state.Negotiation), state.FailedResponses, state.OkResponses)

			state.Transaction.State = persistence.StatusRejected

			err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
			if err != nil {
				log.WithField("transaction", state.Transaction.IDTransaction).Errorf("Commit failed to update transaction %+v", err)
				s.SendMessage(
					RespTransactionRefused+" "+state.Transaction.IDTransaction,
					state.ReplyTo,
					context.Receiver,
				)
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

			log.WithField("transaction", state.Transaction.IDTransaction).Debug("Commit->Rollback")

			return
		}

		log.WithField("transaction", state.Transaction.IDTransaction).Debug("Commit Accepted All")

		state.Transaction.State = persistence.StatusCommitted

		err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
		// FIXME log error
		if err != nil {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)

			log.WithField("transaction", state.Transaction.IDTransaction).Warn("Commit failed to commit transaction")

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

		log.WithField("transaction", state.Transaction.IDTransaction).Info("New Transaction Committed")
		log.WithField("transaction", state.Transaction.IDTransaction).Debug("Commit->End")

		s.UnregisterActor(context.Sender.Name)
		return
	}
}

func RollbackingTransaction(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(TransactionState)
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

			log.WithField("transaction", state.Transaction.IDTransaction).Debugf("Rollback Rejected Some [total: %d, accepted: %d, rejected: %d]", len(state.Negotiation), state.FailedResponses, state.OkResponses)

			s.UnregisterActor(context.Sender.Name)
			return
		}

		log.WithField("transaction", state.Transaction.IDTransaction).Debug("Rollback Accepted All")

		// FIXME
		//rollBackReason := "unknown"

		state.Transaction.State = persistence.StatusRollbacked

		err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
		if err != nil {
			log.WithField("transaction", state.Transaction.IDTransaction).Errorf("Rollback failed to update transaction %+v", err)
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)

			log.WithField("transaction", state.Transaction.IDTransaction).Warn("Rollback failed to rollback transaction")

			s.UnregisterActor(context.Sender.Name)
			return
		}

		s.Metrics.TransactionRollbacked(len(state.Transaction.Transfers))

		s.SendMessage(
			RespTransactionRejected+" "+state.Transaction.IDTransaction+" "+state.Transaction.State,
			state.ReplyTo,
			context.Receiver,
		)

		log.WithField("transaction", state.Transaction.IDTransaction).Info("New Transaction Rollbacked")
		log.WithField("transaction", state.Transaction.IDTransaction).Debug("Rollback->End")

		s.UnregisterActor(context.Sender.Name)
		return
	}
}
