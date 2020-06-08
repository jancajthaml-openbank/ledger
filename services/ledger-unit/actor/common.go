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
	"strings"

	"github.com/jancajthaml-openbank/ledger-unit/model"

	system "github.com/jancajthaml-openbank/actor-system"
	money "gopkg.in/inf.v0"
)

// ProcessRemoteMessage processing of remote message to this wall
func ProcessMessage(s *ActorSystem) system.ProcessMessage {
	return func(msg string, to system.Coordinates, from system.Coordinates) {

		parts := strings.Split(msg, " ")

		var (
			message interface{}
			ref     *system.Envelope
			err     error
		)

		if parts[0] == ReqCreateTransaction {
			if len(parts) > 2 {
				transaction := model.Transaction{
					IDTransaction: parts[1],
				}

				for _, transferPart := range parts[2:] {
					transferParts := strings.Split(transferPart, ";")
					amount, _ := new(money.Dec).SetString(transferParts[5])
					transfer := model.Transfer{
						IDTransfer: transferParts[0],
						Credit: model.Account{
							Tenant: transferParts[1],
							Name:   transferParts[2],
						},
						Debit: model.Account{
							Tenant: transferParts[3],
							Name:   transferParts[4],
						},
						ValueDate: transferParts[7],
						Amount:    amount,
						Currency:  transferParts[6],
					}
					transaction.Transfers = append(transaction.Transfers, transfer)
				}
				message = transaction
				ref, err = NewTransactionActor(s, to.Name)
				if err != nil {
					log.Warnf("Unable to spray from [remote %v -> local %v] : %+v", from, to, msg)
					s.SendMessage(FatalError, from, to)
					return
				}
			}
		} else {
			ref, err = s.ActorOf(to.Name)
			if err != nil {
				return
			}

			switch parts[0] {

			case FatalError:
				message = FatalErrored{
					Account: model.Account{
						Tenant: from.Region[10:],
						Name:   from.Name,
					},
				}

			case PromiseAccepted:
				message = PromiseWasAccepted{
					Account: model.Account{
						Tenant: from.Region[10:],
						Name:   from.Name,
					},
				}

			case PromiseRejected:
				if len(parts) == 2 {
					message = PromiseWasRejected{
						Account: model.Account{
							Tenant: from.Region[10:],
							Name:   from.Name,
						},
						Reason: parts[1],
					}
				}

			case CommitAccepted:
				message = CommitWasAccepted{
					Account: model.Account{
						Tenant: from.Region[10:],
						Name:   from.Name,
					},
				}

			case CommitRejected:
				if len(parts) == 2 {
					message = CommitWasRejected{
						Account: model.Account{
							Tenant: from.Region[10:],
							Name:   from.Name,
						},
						Reason: parts[1],
					}
				}

			case RollbackAccepted:
				message = RollbackWasAccepted{
					Account: model.Account{
						Tenant: from.Region[10:],
						Name:   from.Name,
					},
				}

			case RollbackRejected:
				if len(parts) == 2 {
					message = RollbackWasRejected{
						Account: model.Account{
							Tenant: from.Region[10:],
							Name:   from.Name,
						},
						Reason: parts[1],
					}
				}

			}
		}

		if message == nil {
			log.Warnf("Deserialization of unsuported message [remote %v -> local %v] : %+v", from, to, msg)
			message = FatalError
		}

		ref.Tell(message, to, from)
		return
	}
}

// NewTransactionActor creates new transaction actor
func NewTransactionActor(s *ActorSystem, name string) (*system.Envelope, error) {
	envelope := system.NewEnvelope(name, NewTransactionState())
	err := s.RegisterActor(envelope, InitialTransaction(s))
	if err != nil {
		log.Warnf("Spawning Transaction Actor %s Error unable to register", name)
		return nil, err
	}
	return envelope, nil
}
