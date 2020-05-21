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

	"github.com/jancajthaml-openbank/ledger-unit/model"
	"github.com/jancajthaml-openbank/ledger-unit/persistence"

	system "github.com/jancajthaml-openbank/actor-system"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

func InitialForward(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.ForwardState)

		switch msg := context.Data.(type) {
		case model.TransferForward:

			if state.Ready {
				s.SendMessage(
					TransactionRaceMessage(msg.IDTransaction),
					state.ReplyTo,
					context.Receiver,
				)
				log.Warnf("(FWD) Forward already in progress %s/%s", msg.IDTransaction, msg.IDTransfer)
				return
			}

			var foundTransfer *model.Transfer = nil

			current, err := persistence.LoadTransaction(s.Storage, msg.IDTransaction)
			if err == nil {
				for _, transfer := range current.Transfers {
					if transfer.IDTransfer == msg.IDTransfer {
						foundTransfer = &transfer
						break
					}
				}
			}

			if foundTransfer == nil {
				s.SendMessage(
					TransactionMissingMessage(msg.IDTransaction),
					context.Sender,
					context.Receiver,
				)
				log.Warnf("(FWD) Cannot forward non existant transfer %s/%s", msg.IDTransaction, msg.IDTransfer)
				return
			}

			var transaction model.Transaction

			switch msg.Side {

			case "credit":
				if already, err := persistence.IsTransferForwardedCredit(s.Storage, msg.IDTransaction, msg.IDTransfer); already || err != nil {
					s.SendMessage(
						TransactionRefusedMessage(current),
						context.Sender,
						context.Receiver,
					)
					log.Warnf("(FWD) Transaction %s/%s credit side is already forwarded", msg.IDTransaction, msg.IDTransfer)
					return
				}

				id := xid.New().String()
				transaction = model.Transaction{
					IDTransaction: id,
					Transfers: []model.Transfer{
						{
							IDTransfer: id + "1",
							Credit:     msg.Target,
							Debit:      foundTransfer.Credit,
							ValueDate:  foundTransfer.ValueDate,
							Amount:     foundTransfer.Amount,
							Currency:   foundTransfer.Currency,
						},
					},
				}

			case "debit":
				if already, err := persistence.IsTransferForwardedDebit(s.Storage, msg.IDTransaction, msg.IDTransfer); already || err != nil {
					s.SendMessage(
						TransactionRefusedMessage(current),
						context.Sender,
						context.Receiver,
					)
					log.Warnf("(FWD) Transaction %s/%s debit side is already forwarded", msg.IDTransaction, msg.IDTransfer)
					return
				}

				id := xid.New().String()
				transaction = model.Transaction{
					IDTransaction: id,
					Transfers: []model.Transfer{
						{
							IDTransfer: id + "1",
							Credit:     foundTransfer.Debit,
							Debit:      msg.Target,
							ValueDate:  foundTransfer.ValueDate,
							Amount:     foundTransfer.Amount,
							Currency:   foundTransfer.Currency,
						},
					},
				}

			default:
				s.SendMessage(
					TransactionRefusedMessage(current),
					context.Sender,
					context.Receiver,
				)
				log.Warnf("(FWD) Transaction %s/%s invalid forward side", msg.IDTransaction, msg.IDTransfer)
			}

			if persistence.PersistTransaction(s.Storage, &transaction) == nil {
				s.SendMessage(
					TransactionRaceMessage(state.Forward.IDTransaction),
					context.Sender,
					context.Receiver,
				)
				return
			}


			state.Forward = msg
			state.Prepare(transaction, context.Sender)
			state.ResetMarks()

			context.Self.Become(state, PromisingForward(s))
			s.Metrics.TransactionPromised(1)

			for account, task := range state.Negotiation {
				s.SendMessage(
					PromiseOrderMessage(task),
					system.Coordinates{
						Region: "VaultUnit/" + account.Tenant,
						Name: account.Name,
					},
					context.Receiver,
				)
			}

			log.Infof("~ %v (FWD) Start->Promise", state.Transaction.IDTransaction)

		default:
			s.SendMessage(
				FatalErrorMessage(),
				state.ReplyTo,
				context.Receiver,
			)
			log.Warnf("Invalid message in InitialForward")
			return
		}

	}
}

func PromisingForward(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.ForwardState)

		switch msg := context.Data.(type) {
		case model.PromiseWasAccepted:
			log.Debugf("~ %v (FWD) Promise Accepted %s", state.Transaction.IDTransaction, msg.Account)
		case model.PromiseWasRejected:
			log.Debugf("~ %v (FWD) Promise Rejected %s %s", state.Transaction.IDTransaction, msg.Account, msg.Reason)
		case model.FatalErrored:
			log.Debugf("~ %v (FWD) Promise Errored %s", state.Transaction.IDTransaction, msg.Account)
		default:
			log.Debugf("~ %v (FWD) Promise Invalid Message %+v / %+v", state.Transaction.IDTransaction, reflect.ValueOf(context.Data).Type(), context.Data)
		}

		state.Mark(context.Data)

		if !state.IsNegotiationFinished() {
			context.Self.Become(state, PromisingForward(s))
			return
		}

		if state.OkResponses == 0 {
			s.SendMessage(
				TransactionRefusedMessage(&state.Transaction),
				state.ReplyTo,
				context.Receiver,
			)
			log.Debugf("~ %v (FWD) Promise Rejected All", state.Transaction.IDTransaction)
			s.UnregisterActor(context.Sender.Name)
			return
		} else if state.FailedResponses > 0 {
			log.Debugf("~ %v (FWD) Promise Rejected Some [total: %d, accepted: %d, rejected : %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)

			state.Transaction.State = model.StatusRejected

			if persistence.UpdateTransaction(s.Storage, &state.Transaction) == nil {
				s.SendMessage(
					TransactionRefusedMessage(&state.Transaction),
					state.ReplyTo,
					context.Receiver,
				)
				log.Warnf("~ %v (FWD) Promise failed to reject transaction", state.Transaction.IDTransaction)
				s.UnregisterActor(context.Sender.Name)
				return
			}

			log.Infof("~ %v (FWD) Promise->Rollback", state.Transaction.IDTransaction)
			state.ResetMarks()
			context.Self.Become(state, RollbackingForward(s))

			for account, task := range state.Negotiation {
				s.SendMessage(
					RollbackOrderMessage(task),
					system.Coordinates{
						Region: "VaultUnit/" + account.Tenant,
						Name: account.Name,
					},
					context.Receiver,
				)
			}

			return
		}

		log.Debugf("~ %v (FWD) Promise Accepted All", state.Transaction.IDTransaction)

		state.Transaction.State = model.StatusAccepted

		// FIXME possible null here
		if persistence.UpdateTransaction(s.Storage, &state.Transaction) == nil {
			s.SendMessage(
				TransactionRefusedMessage(&state.Transaction),
				state.ReplyTo,
				context.Receiver,
			)
			log.Warnf("~ %v (FWD) Promise failed to accept transaction", state.Transaction.IDTransaction)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		state.ResetMarks()
		context.Self.Become(state, CommitingForward(s))

		for account, task := range state.Negotiation {
			s.SendMessage(
				CommitOrderMessage(task),
				system.Coordinates{
					Region: "VaultUnit/" + account.Tenant,
					Name: account.Name,
				},
				context.Receiver,
			)
		}

		log.Infof("~ %v (FWD) Promise->Commit", state.Transaction.IDTransaction)
		return
	}
}

func CommitingForward(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.ForwardState)

		switch msg := context.Data.(type) {
		case model.CommitWasAccepted:
			log.Debugf("~ %v (FWD) Commit Accepted %s", state.Transaction.IDTransaction, msg.Account)
		case model.CommitWasRejected:
			log.Debugf("~ %v (FWD) Commit Rejected %s %s", state.Transaction.IDTransaction, msg.Account, msg.Reason)
		case model.FatalErrored:
			log.Debugf("~ %v (FWD) Commit Errored %s", state.Transaction.IDTransaction, msg.Account)
		default:
			log.Debugf("~ %v (FWD) Commit Invalid Message %+v / %+v", state.Transaction.IDTransaction, reflect.ValueOf(context.Data).Type(), context.Data)
		}
		state.Mark(context.Data)

		if !state.IsNegotiationFinished() {
			context.Self.Become(state, CommitingForward(s))
			return
		}

		if state.FailedResponses > 0 {
			log.Debugf("~ %v (FWD) Commit Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)

			state.Transaction.State = model.StatusRejected

			if persistence.UpdateTransaction(s.Storage, &state.Transaction) == nil {
				s.SendMessage(
					TransactionRefusedMessage(&state.Transaction),
					state.ReplyTo,
					context.Receiver,
				)
				log.Warnf("~ %v (FWD) Commit failed to reject transaction", state.Transaction.IDTransaction)
				s.UnregisterActor(context.Sender.Name)
				return
			}

			state.ResetMarks()
			context.Self.Become(state, RollbackingForward(s))

			for account, task := range state.Negotiation {
				s.SendMessage(
					RollbackOrderMessage(task),
					system.Coordinates{
						Region: "VaultUnit/" + account.Tenant,
						Name: account.Name,
					},
					context.Receiver,
				)
			}

			log.Infof("~ %v (FWD) Commit->Rollback", state.Transaction.IDTransaction)
			return
		}

		log.Debugf("~ %v (FWD) Commit Accepted All", state.Transaction.IDTransaction)

		state.Transaction.State = model.StatusCommitted

		if persistence.UpdateTransaction(s.Storage, &state.Transaction) == nil {
			s.SendMessage(
				TransactionRefusedMessage(&state.Transaction),
				state.ReplyTo,
				context.Receiver,
			)
			log.Warnf("~ %v (FWD) Commit failed to commit transaction", state.Transaction.IDTransaction)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		var transfers []string
		for _, transfer := range state.Transaction.Transfers {
			transfers = append(transfers, transfer.IDTransfer)
		}

		state.ResetMarks()
		context.Self.Become(state, AcceptingForward(s))

		s.Metrics.TransactionCommitted(1)

		s.SendMessage(
			TransactionProcessedMessage(&state.Transaction),
			state.ReplyTo,
			context.Receiver,
		)

		log.Infof("~ %v (FWD) Commit->Accept", state.Transaction.IDTransaction)

		return
	}
}

func RollbackingForward(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.ForwardState)

		switch msg := context.Data.(type) {
		case model.RollbackWasAccepted:
			log.Debugf("~ %v (FWD) Rollback Accepted %s", state.Transaction.IDTransaction, msg.Account)
		case model.RollbackWasRejected:
			log.Debugf("~ %v (FWD) Rollback Rejected %s %s", state.Transaction.IDTransaction, msg.Account, msg.Reason)
		case model.FatalErrored:
			log.Debugf("~ %v (FWD) Rollback Errored %s", state.Transaction.IDTransaction, msg.Account)
		default:
			log.Debugf("~ %v (FWD) Rollback Invalid Message %+v / %+v", state.Transaction.IDTransaction, reflect.ValueOf(context.Data).Type(), context.Data)
		}
		state.Mark(context.Data)

		if !state.IsNegotiationFinished() {
			context.Self.Become(state, RollbackingForward(s))
			return
		}

		if state.FailedResponses > 0 {
			s.SendMessage(
				TransactionRefusedMessage(&state.Transaction),
				state.ReplyTo,
				context.Receiver,
			)
			log.Debugf("~ %v (FWD) Rollback Rejected Rejected [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		log.Debugf("~ %v (FWD) Rollback Accepted All", state.Transaction.IDTransaction)

		// FIXME
		//rollBackReason := "unknown"

		state.Transaction.State = model.StatusRollbacked

		if persistence.UpdateTransaction(s.Storage, &state.Transaction) == nil {
			log.Warnf("~ %v (FWD) Rollback failed to rollback transaction", state.Transaction.IDTransaction)
			s.SendMessage(
				TransactionRefusedMessage(&state.Transaction),
				state.ReplyTo,
				context.Receiver,
			)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		s.Metrics.TransactionRollbacked(1)

		s.SendMessage(
			TransactionRejectedMessage(&state.Transaction),
			state.ReplyTo,
			context.Receiver,
		)
		log.Infof("~ %v (FWD) Rollback->End", state.Transaction.IDTransaction)
		s.UnregisterActor(context.Sender.Name)
		return
	}
}

func AcceptingForward(s *ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.ForwardState)

		var ok bool
		switch state.Forward.Side {

		case "credit":
			ok = persistence.AcceptForwardCredit(s.Storage, state.Forward.Target.Tenant, state.Transaction.IDTransaction, state.Transaction.Transfers[0].IDTransfer, state.Forward.IDTransaction, state.Forward.IDTransfer) != nil

		case "debit":
			ok = persistence.AcceptForwardDebit(s.Storage, state.Forward.Target.Tenant, state.Transaction.IDTransaction, state.Transaction.Transfers[0].IDTransfer, state.Forward.IDTransaction, state.Forward.IDTransfer) != nil

		}

		if !ok {
			s.SendMessage(
				TransactionRejectedMessage(&state.Transaction),
				state.ReplyTo,
				context.Receiver,
			)
			log.Warnf("~ %v (FWD) Accept failed to accept forward", state.Transaction.IDTransaction)
			s.UnregisterActor(context.Sender.Name)
			return
		}

		s.Metrics.TransactionForwarded(1)

		s.SendMessage(
			TransactionProcessedMessage(&state.Transaction),
			state.ReplyTo,
			context.Receiver,
		)
		log.Infof("~ %v (FWD) Accept->End", state.Transaction.IDTransaction)
		s.UnregisterActor(context.Sender.Name)
		return
	}
}
