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
	"strings"

	"github.com/jancajthaml-openbank/ledger-unit/model"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

// ProcessRemoteMessage processing of remote message to this wall
func ProcessMessage(s *ActorSystem) system.ProcessMessage {
	return func(msg string, to system.Coordinates, from system.Coordinates) {

		defer func() {
			if r := recover(); r != nil {
				log.Errorf("procesRemoteMessage recovered in [remote %v -> local %v] : %+v", from, to, r)
			}
		}()

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
				ref, err = spawnTransactionActor(s, to.Name)
				if err != nil {
					log.Warnf("Unable to spray from [remote %v -> local %v] : %+v", from, to, msg)
					s.SendMessage(FatalErrorMessage(), from, to)
					return
				}
			}
		} else if parts[0] == ReqForwardTransfer {
			if len(parts) == 5 {
				targetParts := strings.Split(parts[4], ";")
				message = model.TransferForward{
					IDTransaction: parts[1],
					IDTransfer:    parts[2],
					Side:          parts[3],
					Target: model.Account{
						Tenant: targetParts[0],
						Name:   targetParts[1],
					},
				}

				ref, err = spawnForwardActor(s, to.Name)
				if err != nil {
					log.Warnf("Unable to spray from [remote %v -> local %v] : %+v", from, to, msg)
					s.SendMessage(FatalErrorMessage(), from, to)
					return
				}
			}
		} else {
			ref, err = s.ActorOf(to.Name)
			if err != nil {
				log.Warnf("Deadletter received [remote %v -> local %v] : %+v", from, to, msg)
				return
			}

			switch parts[0] {

			case FatalError:
				message = model.FatalErrored{
					Account: model.Account{
						Tenant: from.Region[10:],
						Name:   from.Name,
					},
				}

			case PromiseAccepted:
				message = model.PromiseWasAccepted{
					Account: model.Account{
						Tenant: from.Region[10:],
						Name:   from.Name,
					},
				}

			case PromiseRejected:
				if len(parts) == 2 {
					message = model.PromiseWasRejected{
						Account: model.Account{
							Tenant: from.Region[10:],
							Name:   from.Name,
						},
						Reason: parts[1],
					}
				}

			case CommitAccepted:
				message = model.CommitWasAccepted{
					Account: model.Account{
						Tenant: from.Region[10:],
						Name:   from.Name,
					},
				}

			case CommitRejected:
				if len(parts) == 2 {
					message = model.CommitWasRejected{
						Account: model.Account{
							Tenant: from.Region[10:],
							Name:   from.Name,
						},
						Reason: parts[1],
					}
				}

			case RollbackAccepted:
				message = model.RollbackWasAccepted{
					Account: model.Account{
						Tenant: from.Region[10:],
						Name:   from.Name,
					},
				}

			case RollbackRejected:
				if len(parts) == 2 {
					message = model.RollbackWasRejected{
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

func spawnForwardActor(s *ActorSystem, name string) (*system.Envelope, error) {
	envelope := system.NewEnvelope(name, model.NewForwardState())

	err := s.RegisterActor(envelope, InitialForward(s))
	if err != nil {
		log.Warnf("%s ~ Spawning Forward Actor Error unable to register", name)
		return nil, err
	}

	log.Debugf("%s ~ Forward Actor Spawned", name)
	return envelope, nil
}

func spawnTransactionActor(s *ActorSystem, name string) (*system.Envelope, error) {
	envelope := system.NewEnvelope(name, model.NewTransactionState())

	err := s.RegisterActor(envelope, InitialTransaction(s))
	if err != nil {
		log.Warnf("%s ~ Spawning Transaction Actor Error unable to register", name)
		return nil, err
	}

	log.Debugf("%s ~ Transaction Actor Spawned", name)
	return envelope, nil
}
