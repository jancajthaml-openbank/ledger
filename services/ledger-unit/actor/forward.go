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
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

func InitialForward(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.ForwardState)

		switch msg := context.Data.(type) {
		case model.TransferForward:

			if state.Ready {
				log.Warnf("(FWD) Forward already in progress %s/%s", msg.IDTransaction, msg.IDTransfer)
				s.SendRemote(context.Sender.Region, TransactionRaceMessage(context.Receiver.Name, context.Sender.Name, msg.IDTransaction))
				return
			}

			existing := persistence.LoadTransfer(s.Storage, msg.IDTransaction, msg.IDTransfer)
			if existing == nil {
				log.Warnf("(FWD) Cannot forward non existant transfer %s/%s", msg.IDTransaction, msg.IDTransfer)
				s.SendRemote(context.Sender.Region, TransactionMissingMessage(context.Receiver.Name, context.Sender.Name, msg.IDTransaction))
				return
			}

			var transaction model.Transaction

			switch msg.Side {

			case "credit":
				if already, err := persistence.IsTransferForwardedCredit(s.Storage, msg.IDTransaction, msg.IDTransfer); already || err != nil {
					log.Warnf("(FWD) Transaction %s/%s credit side is already forwarded", msg.IDTransaction, msg.IDTransfer)
					s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, msg.IDTransaction))
					return
				}

				id := xid.New().String()
				transaction = model.Transaction{
					IDTransaction: id,
					Transfers: []model.Transfer{
						{
							IDTransfer: id + "1",
							Credit:     msg.Target,
							Debit:      existing.Credit,
							ValueDate:  existing.ValueDate,
							Amount:     existing.Amount,
							Currency:   existing.Currency,
						},
					},
				}

			case "debit":
				if already, err := persistence.IsTransferForwardedDebit(s.Storage, msg.IDTransaction, msg.IDTransfer); already || err != nil {
					log.Warnf("(FWD) Transaction %s/%s debit side is already forwarded", msg.IDTransaction, msg.IDTransfer)
					s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, msg.IDTransaction))
					return
				}

				id := xid.New().String()
				transaction = model.Transaction{
					IDTransaction: id,
					Transfers: []model.Transfer{
						{
							IDTransfer: id + "1",
							Credit:     existing.Debit,
							Debit:      msg.Target,
							ValueDate:  existing.ValueDate,
							Amount:     existing.Amount,
							Currency:   existing.Currency,
						},
					},
				}

			default:
				log.Warnf("(FWD) Transaction %s/%s invalid forward side", msg.IDTransaction, msg.IDTransfer)
				s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, msg.IDTransaction))
			}

			if persistence.PersistTransaction(s.Storage, &transaction) == nil {
				s.SendRemote(context.Sender.Region, TransactionRaceMessage(context.Receiver.Name, context.Sender.Name, state.Forward.IDTransaction))
				return
			}

			state.Forward = msg
			state.Prepare(transaction)

			log.Infof("~ %v (FWD) Start->Promise", state.Transaction.IDTransaction)
			state.ResetMarks()
			context.Receiver.Become(state, PromisingForward(s))
			s.Metrics.TransactionPromised(1)

			for account, task := range state.Negotiation {
				s.SendRemote("VaultUnit/"+account.Tenant, PromiseOrderMessage(context.Receiver.Name, account.Name, task))
			}

		default:
			log.Warnf("Invalid message in InitialForward")
			s.SendRemote(context.Sender.Region, FatalErrorMessage(context.Receiver.Name, context.Sender.Name))
			return
		}

	}
}

func PromisingForward(s *daemon.ActorSystem) func(interface{}, system.Context) {
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
			context.Receiver.Become(state, PromisingForward(s))
			return
		}

		if state.OkResponses == 0 {
			log.Debugf("~ %v (FWD) Promise Rejected All", state.Transaction.IDTransaction)
			s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
			return
		} else if state.FailedResponses > 0 {
			log.Debugf("~ %v (FWD) Promise Rejected Some [total: %d, accepted: %d, rejected : %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)
			if !persistence.RejectTransaction(s.Storage, state.Transaction.IDTransaction) {
				log.Warnf("~ %v (FWD) Promise failed to reject transaction", state.Transaction.IDTransaction)
				s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
				return
			}

			log.Infof("~ %v (FWD) Promise->Rollback", state.Transaction.IDTransaction)
			state.ResetMarks()
			context.Receiver.Become(state, RollbackingForward(s))

			for account, task := range state.Negotiation {
				s.SendRemote("VaultUnit/"+account.Tenant, RollbackOrderMessage(context.Receiver.Name, account.Name, task))
			}

			return
		}

		log.Debugf("~ %v (FWD) Promise Accepted All", state.Transaction.IDTransaction)

		// FIXME possible null here
		if !persistence.AcceptTransaction(s.Storage, state.Transaction.IDTransaction) {
			log.Warnf("~ %v (FWD) Promise failed to accept transaction", state.Transaction.IDTransaction)
			s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
			return
		}

		log.Infof("~ %v (FWD) Promise->Commit", state.Transaction.IDTransaction)
		state.ResetMarks()
		context.Receiver.Become(state, CommitingForward(s))

		for account, task := range state.Negotiation {
			s.SendRemote("VaultUnit/"+account.Tenant, CommitOrderMessage(context.Receiver.Name, account.Name, task))
		}
		return
	}
}

func CommitingForward(s *daemon.ActorSystem) func(interface{}, system.Context) {
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
			context.Receiver.Become(state, CommitingForward(s))
			return
		}

		if state.FailedResponses > 0 {
			log.Debugf("~ %v (FWD) Commit Rejected Some [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)
			if !persistence.RejectTransaction(s.Storage, state.Transaction.IDTransaction) {
				log.Warnf("~ %v (FWD) Commit failed to reject transaction", state.Transaction.IDTransaction)
				s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
				return
			}

			log.Infof("~ %v (FWD) Commit->Rollback", state.Transaction.IDTransaction)
			state.ResetMarks()
			context.Receiver.Become(state, RollbackingForward(s))

			for account, task := range state.Negotiation {
				s.SendRemote("VaultUnit/"+account.Tenant, RollbackOrderMessage(context.Receiver.Name, account.Name, task))
			}

			return
		}

		log.Debugf("~ %v (FWD) Commit Accepted All", state.Transaction.IDTransaction)

		if !persistence.CommitTransaction(s.Storage, state.Transaction.IDTransaction) {
			log.Warnf("~ %v (FWD) Commit failed to commit transaction", state.Transaction.IDTransaction)
			s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
			return
		}

		var transfers []string
		for _, transfer := range state.Transaction.Transfers {
			transfers = append(transfers, transfer.IDTransfer)
		}

		log.Infof("~ %v (FWD) Commit->Accept", state.Transaction.IDTransaction)

		s.SendRemote(context.Sender.Region, TransactionProcessedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))

		s.Metrics.TransactionCommitted(1)
		state.ResetMarks()
		context.Receiver.Become(state, AcceptingForward(s))
		return
	}
}

func RollbackingForward(s *daemon.ActorSystem) func(interface{}, system.Context) {
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
			context.Receiver.Become(state, RollbackingForward(s))
			return
		}

		if state.FailedResponses > 0 {
			log.Debugf("~ %v (FWD) Rollback Rejected Rejected [total: %d, accepted: %d, rejected: %d]", state.Transaction.IDTransaction, len(state.Negotiation), state.FailedResponses, state.OkResponses)
			s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
			return
		}

		log.Debugf("~ %v (FWD) Rollback Accepted All", state.Transaction.IDTransaction)

		rollBackReason := "unknown"

		if !persistence.RollbackTransaction(s.Storage, state.Transaction.IDTransaction, rollBackReason) {
			log.Warnf("~ %v (FWD) Rollback failed to rollback transaction", state.Transaction.IDTransaction)
			s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
			return
		}

		log.Infof("~ %v (FWD) Rollback->End", state.Transaction.IDTransaction)
		s.SendRemote(context.Sender.Region, TransactionRejectedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction, rollBackReason))

		s.Metrics.TransactionRollbacked(1)
		s.UnregisterActor(context.Sender.Name)
		return
	}
}

func AcceptingForward(s *daemon.ActorSystem) func(interface{}, system.Context) {
	return func(t_state interface{}, context system.Context) {
		state := t_state.(model.ForwardState)

		var ok bool
		switch state.Forward.Side {

		case "credit":
			ok = persistence.AcceptForwardCredit(s.Storage, state.Forward.Target.Tenant, state.Transaction.IDTransaction, state.Transaction.Transfers[0].IDTransfer, state.Forward.IDTransaction, state.Forward.IDTransfer)

		case "debit":
			ok = persistence.AcceptForwardDebit(s.Storage, state.Forward.Target.Tenant, state.Transaction.IDTransaction, state.Transaction.Transfers[0].IDTransfer, state.Forward.IDTransaction, state.Forward.IDTransfer)

		}

		if !ok {
			log.Warnf("~ %v (FWD) Accept failed to accept forward", state.Transaction.IDTransaction)
			s.SendRemote(context.Sender.Region, TransactionRefusedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
			return
		}

		log.Infof("~ %v (FWD) Accept->End", state.Transaction.IDTransaction)

		s.SendRemote(context.Sender.Region, TransactionProcessedMessage(context.Receiver.Name, context.Sender.Name, state.Transaction.IDTransaction))
		s.Metrics.TransactionForwarded(1)
		s.UnregisterActor(context.Sender.Name)
		return
	}
}
