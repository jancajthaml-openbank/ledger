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

// InitialTransaction represents initial transaction state
func InitialTransaction(s *System) func(interface{}, system.Context) {
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
				log.Warn().Msgf("%s/Initial already in progress", state.Transaction.IDTransaction)
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

		log.Debug().Msgf("%s/Initial -> %s/Promise", state.Transaction.IDTransaction, state.Transaction.IDTransaction)
	}
}

// PromisingTransaction represents transaction in promising state
func PromisingTransaction(s *System) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(TransactionState)

		accountRetry := state.Mark(context.Data)

		if accountRetry != nil {
			log.Debug().Msgf("%s/Promise Bounced for %v", state.Transaction.IDTransaction, accountRetry)

			for account, task := range state.Negotiation {
				if !(accountRetry.Tenant == account.Tenant && accountRetry.Name != account.Name) {
					continue
				}
				s.SendMessage(
					CommitOrder+" "+task,
					system.Coordinates{
						Region: "VaultUnit/" + account.Tenant,
						Name:   account.Name,
					},
					context.Receiver,
				)
			}
			return
		}

		if !state.IsNegotiationFinished() {
			context.Self.Become(state, PromisingTransaction(s))
			return
		}

		if state.FailedResponses > 0 {
			state.Transaction.State = persistence.StatusRejected
			err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
			if err != nil {
				log.Error().Msgf("%s/Promise failed to update transaction %+v", state.Transaction.IDTransaction, err)
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
			log.Debug().Msgf("%s/Promise Rejected All", state.Transaction.IDTransaction)
			return
		}

		if state.FailedResponses > 0 {
			log.Debug().Msgf("%s/Promise Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)
			log.Debug().Msgf("%s/Promise -> %s/Rollback", state.Transaction.IDTransaction, state.Transaction.IDTransaction)

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

		log.Debug().Msgf("%s/Promise Accepted All", state.Transaction.IDTransaction)

		state.Transaction.State = persistence.StatusAccepted

		err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
		if err != nil {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)

			log.Warn().Msgf("%s/Promise failed to accept transaction", state.Transaction.IDTransaction)

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
		log.Debug().Msgf("%s/Promise -> %s/Commit", state.Transaction.IDTransaction, state.Transaction.IDTransaction)
		return
	}
}

// CommitingTransaction represents transaction in committing state
func CommitingTransaction(s *System) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(TransactionState)
		state.Mark(context.Data)
		if !state.IsNegotiationFinished() {
			context.Self.Become(state, CommitingTransaction(s))
			return
		}

		if state.FailedResponses > 0 {
			log.Debug().Msgf("%s/Commit Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)

			state.Transaction.State = persistence.StatusRejected

			err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
			if err != nil {
				log.Error().Msgf("%s/Commit failed to update transaction %+v", state.Transaction.IDTransaction, err)
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

			log.Debug().Msgf("%s/Commit -> %s/Rollback", state.Transaction.IDTransaction, state.Transaction.IDTransaction)

			return
		}

		log.Debug().Msgf("%s/Commit Accepted All", state.Transaction.IDTransaction)

		state.Transaction.State = persistence.StatusCommitted

		err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
		// FIXME log error
		if err != nil {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)

			log.Warn().Msgf("%s/Commit failed to commit transaction", state.Transaction.IDTransaction)

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

		log.Info().Msgf("New Transaction %s Committed", state.Transaction.IDTransaction)
		log.Debug().Msgf("%s/Commit -> Unregister", state.Transaction.IDTransaction)

		s.UnregisterActor(context.Sender.Name)
		return
	}
}

// RollbackingTransaction represents transaction in rollbacking state
func RollbackingTransaction(s *System) func(interface{}, system.Context) {
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

			log.Debug().Msgf("%s/Rollback Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)

			s.UnregisterActor(context.Sender.Name)
			return
		}

		log.Debug().Msgf("%s/Rollback Accepted All", state.Transaction.IDTransaction)

		// FIXME
		//rollBackReason := "unknown"

		state.Transaction.State = persistence.StatusRollbacked

		err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
		if err != nil {
			log.Error().Msgf("%s/Rollback failed to update transaction %+v", state.Transaction.IDTransaction, err)
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)

			log.Warn().Msgf("%s/Rollback failed to rollback transaction", state.Transaction.IDTransaction)

			s.UnregisterActor(context.Sender.Name)
			return
		}

		s.Metrics.TransactionRollbacked(len(state.Transaction.Transfers))

		s.SendMessage(
			RespTransactionRejected+" "+state.Transaction.IDTransaction+" "+state.Transaction.State,
			state.ReplyTo,
			context.Receiver,
		)

		log.Info().Msgf("New Transaction %s Rollbacked", state.Transaction.IDTransaction)
		log.Debug().Msgf("%s/Rollback -> Unregister", state.Transaction.IDTransaction)

		s.UnregisterActor(context.Sender.Name)
		return
	}
}
