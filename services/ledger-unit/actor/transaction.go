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
	"github.com/jancajthaml-openbank/ledger-unit/persistence"

	system "github.com/jancajthaml-openbank/actor-system"
)

// InitialTransaction represents initial transaction state
func InitialTransaction(s *System, state TransactionState) system.ReceiverFunction {
	return func(context system.Context) system.ReceiverFunction {

		msg, ok := context.Data.(model.Transaction)
		if !ok {
			s.SendMessage(FatalError, state.ReplyTo, context.Receiver)
			s.UnregisterActor(context.Sender.Name)
			return InitialTransaction(s, state)
		}

		msg.State = model.StatusNew

		if err := persistence.CreateTransaction(s.Storage, &msg); err != nil {

			current, err := persistence.LoadTransaction(s.Storage, msg.IDTransaction)
			if err != nil {
				log.Warn().Msgf("%s/Initial Conflict storage bounce", msg.IDTransaction)
				s.SendMessage(
					RespTransactionRace+" "+msg.IDTransaction,
					context.Sender,
					context.Receiver,
				)
				s.UnregisterActor(context.Sender.Name)
				return InitialTransaction(s, state)
			}

			switch current.State {

			case model.StatusCommitted, model.StatusRollbacked:
				if msg.IsSameAs(current) {

					var reply string
					if current.State == model.StatusCommitted {
						log.Debug().Msgf("%s/Initial Conflict already committed", current.IDTransaction)
						reply = RespCreateTransaction + " " + current.IDTransaction
					} else {
						log.Debug().Msgf("%s/Initial Conflict already rejected", current.IDTransaction)
						reply = RespTransactionRejected + " " + current.IDTransaction + " " + current.State
					}

					log.Debug().Msgf("%s/Initial -> Terminal", current.IDTransaction)

					s.SendMessage(
						reply,
						context.Sender,
						context.Receiver,
					)

					s.UnregisterActor(context.Sender.Name)
					return InitialTransaction(s, state)
				} else {
					log.Debug().Msgf("%s/Initial Conflict duplicate. Existing %+v vs requested %+v", current.IDTransaction, current, msg)

					log.Debug().Msgf("%s/Initial -> Terminal", current.IDTransaction)

					s.SendMessage(
						RespTransactionDuplicate+" "+current.IDTransaction+" "+current.State,
						context.Sender,
						context.Receiver,
					)

					s.UnregisterActor(context.Sender.Name)
					return InitialTransaction(s, state)
				}

			default:
				log.Debug().Msgf("%s/Initial Conflict status bounce", current.IDTransaction)

				s.SendMessage(
					RespTransactionRace+" "+current.IDTransaction,
					context.Sender,
					context.Receiver,
				)

				s.UnregisterActor(context.Sender.Name)
				return InitialTransaction(s, state)
			}
		}

		state.PrepareNewForTransaction(msg, context.Sender)

		s.Metrics.TransactionPromised(len(state.Transaction.Transfers))

		log.Debug().Msgf("%s/Initial -> %s/Promise", state.Transaction.IDTransaction, state.Transaction.IDTransaction)

		state.ChangeStage(PROMISE)

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

		return PromisingTransaction(s, state)
	}
}

// PromisingTransaction represents transaction in promising state
func PromisingTransaction(s *System, state TransactionState) system.ReceiverFunction {
	return func(context system.Context) system.ReceiverFunction {
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
			return PromisingTransaction(s, state)
		}

		if !state.IsNegotiationFinished() {
			return PromisingTransaction(s, state)
		}

		if state.FailedResponses > 0 {
			state.Transaction.State = model.StatusRejected
			err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
			if err != nil {
				log.Error().Err(err).Msgf("%s/Promise failed to update transaction", state.Transaction.IDTransaction)
				s.SendMessage(
					RespTransactionRefused+" "+state.Transaction.IDTransaction,
					state.ReplyTo,
					context.Receiver,
				)
				s.UnregisterActor(context.Sender.Name)
				return PromisingTransaction(s, state)
			}
		}

		if state.OkResponses == 0 {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)
			log.Debug().Msgf("%s/Promise Rejected All", state.Transaction.IDTransaction)
			return PromisingTransaction(s, state)
		}

		if state.FailedResponses > 0 {
			log.Debug().Msgf("%s/Promise Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)

			log.Debug().Msgf("%s/Promise -> %s/Rollback", state.Transaction.IDTransaction, state.Transaction.IDTransaction)

			state.ChangeStage(ROLLBACK)

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

			return RollbackingTransaction(s, state)
		}

		log.Debug().Msgf("%s/Promise Accepted All", state.Transaction.IDTransaction)

		state.Transaction.State = model.StatusAccepted

		err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
		if err != nil {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)

			log.Warn().Msgf("%s/Promise failed to accept transaction", state.Transaction.IDTransaction)

			s.UnregisterActor(context.Sender.Name)
			return PromisingTransaction(s, state)
		}

		log.Debug().Msgf("%s/Promise -> %s/Commit", state.Transaction.IDTransaction, state.Transaction.IDTransaction)

		state.ChangeStage(COMMIT)

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

		return CommitingTransaction(s, state)
	}
}

// CommitingTransaction represents transaction in committing state
func CommitingTransaction(s *System, state TransactionState) system.ReceiverFunction {
	return func(context system.Context) system.ReceiverFunction {
		state.Mark(context.Data)

		if !state.IsNegotiationFinished() {
			return CommitingTransaction(s, state)
		}

		if state.FailedResponses > 0 {
			log.Debug().Msgf("%s/Commit Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)

			state.Transaction.State = model.StatusRejected

			err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
			if err != nil {
				log.Error().Err(err).Msgf("%s/Commit failed to update transaction", state.Transaction.IDTransaction)
				s.SendMessage(
					RespTransactionRefused+" "+state.Transaction.IDTransaction,
					state.ReplyTo,
					context.Receiver,
				)
				s.UnregisterActor(context.Sender.Name)
				return CommitingTransaction(s, state)
			}

			log.Debug().Msgf("%s/Commit -> %s/Rollback", state.Transaction.IDTransaction, state.Transaction.IDTransaction)

			state.ChangeStage(ROLLBACK)

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

			return RollbackingTransaction(s, state)
		}

		log.Debug().Msgf("%s/Commit Accepted All", state.Transaction.IDTransaction)

		state.Transaction.State = model.StatusCommitted

		err := persistence.UpdateTransaction(s.Storage, &state.Transaction)

		if err != nil {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)

			log.Warn().Msgf("%s/Commit failed to commit transaction", state.Transaction.IDTransaction)

			s.UnregisterActor(context.Sender.Name)
			return CommitingTransaction(s, state)
		}

		s.SendMessage(
			RespCreateTransaction+" "+state.Transaction.IDTransaction,
			state.ReplyTo,
			context.Receiver,
		)

		s.Metrics.TransactionCommitted(len(state.Transaction.Transfers))

		log.Info().Msgf("New Transaction %s Committed", state.Transaction.IDTransaction)
		log.Debug().Msgf("%s/Commit -> Terminal", state.Transaction.IDTransaction)

		s.UnregisterActor(context.Sender.Name)
		return CommitingTransaction(s, state)
	}
}

// RollbackingTransaction represents transaction in rollbacking state
func RollbackingTransaction(s *System, state TransactionState) system.ReceiverFunction {
	return func(context system.Context) system.ReceiverFunction {
		state.Mark(context.Data)

		if !state.IsNegotiationFinished() {
			return RollbackingTransaction(s, state)
		}

		if state.FailedResponses > 0 {
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)

			log.Debug().Msgf("%s/Rollback Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)

			s.UnregisterActor(context.Sender.Name)
			return RollbackingTransaction(s, state)
		}

		log.Debug().Msgf("%s/Rollback Accepted All", state.Transaction.IDTransaction)

		// FIXME
		//rollBackReason := "unknown"

		state.Transaction.State = model.StatusRollbacked

		err := persistence.UpdateTransaction(s.Storage, &state.Transaction)
		if err != nil {
			log.Error().Err(err).Msgf("%s/Rollback failed to update transaction", state.Transaction.IDTransaction)
			s.SendMessage(
				RespTransactionRefused+" "+state.Transaction.IDTransaction,
				state.ReplyTo,
				context.Receiver,
			)

			log.Warn().Msgf("%s/Rollback failed to rollback transaction", state.Transaction.IDTransaction)

			s.UnregisterActor(context.Sender.Name)
			return RollbackingTransaction(s, state)
		}

		s.Metrics.TransactionRollbacked(len(state.Transaction.Transfers))

		s.SendMessage(
			RespTransactionRejected+" "+state.Transaction.IDTransaction+" "+state.Transaction.State,
			state.ReplyTo,
			context.Receiver,
		)

		log.Info().Msgf("New Transaction %s Rollbacked", state.Transaction.IDTransaction)
		log.Debug().Msgf("%s/Rollback -> Terminal", state.Transaction.IDTransaction)

		s.UnregisterActor(context.Sender.Name)
		return RollbackingTransaction(s, state)
	}
}
