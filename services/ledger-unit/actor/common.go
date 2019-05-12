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
	"fmt"
	"strings"

	"github.com/jancajthaml-openbank/ledger-unit/daemon"
	"github.com/jancajthaml-openbank/ledger-unit/model"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
	money "gopkg.in/inf.v0"
)

var nilCoordinates = system.Coordinates{}

func asEnvelopes(s *daemon.ActorSystem, msg string) (system.Coordinates, system.Coordinates, []string, error) {
	parts := strings.Split(msg, " ")

	if len(parts) < 5 {
		return nilCoordinates, nilCoordinates, nil, fmt.Errorf("invalid message received %+v", parts)
	}

	recieverRegion, senderRegion, receiverName, senderName := parts[0], parts[1], parts[2], parts[3]

	from := system.Coordinates{
		Name:   senderName,
		Region: senderRegion,
	}

	to := system.Coordinates{
		Name:   receiverName,
		Region: recieverRegion,
	}

	return from, to, parts, nil
}

// ProcessRemoteMessage processing of remote message to this wall
func ProcessRemoteMessage(s *daemon.ActorSystem) system.ProcessRemoteMessage {
	return func(msg string) {
		from, to, parts, err := asEnvelopes(s, msg)
		if err != nil {
			log.Warn(err.Error())
			return
		}

		defer func() {
			if r := recover(); r != nil {
				log.Errorf("procesRemoteMessage recovered in [remote %v -> local %v] : %+v", from, to, r)
			}
		}()

		var (
			message interface{}
			ref     *system.Envelope
		)

		if parts[4] == ReqCreateTransaction {
			if len(parts) > 6 {
				transaction := model.Transaction{
					IDTransaction: parts[5],
				}

				for _, transferPart := range parts[6:] {
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
					log.Warnf("Unable to spray from [remote %v -> local %v] : %+v", from, to, parts[3:])
					s.SendRemote(FatalErrorMessage(system.Context{
						Receiver: to,
						Sender:   from,
					}))
					return
				}
			}
		} else if parts[4] == ReqForwardTransfer {
			if len(parts) == 9 {
				targetParts := strings.Split(parts[8], ";")
				message = model.TransferForward{
					IDTransaction: parts[5],
					IDTransfer:    parts[6],
					Side:          parts[7],
					Target: model.Account{
						Tenant: targetParts[0],
						Name:   targetParts[1],
					},
				}

				ref, err = spawnForwardActor(s, to.Name)
				if err != nil {
					log.Warnf("Unable to spray from [remote %v -> local %v] : %+v", from, to, msg)
					s.SendRemote(FatalErrorMessage(system.Context{
						Receiver: to,
						Sender:   from,
					}))
					return
				}
			}
		} else {
			ref, err = s.ActorOf(to.Name)
			if err != nil {
				log.Warnf("Deadletter received [remote %v -> local %v] : %+v", from, to, msg)
				return
			}

			switch parts[4] {

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
				if len(parts) == 6 {
					message = model.PromiseWasRejected{
						Account: model.Account{
							Tenant: from.Region[10:],
							Name:   from.Name,
						},
						Reason: parts[5],
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
				if len(parts) == 6 {
					message = model.CommitWasRejected{
						Account: model.Account{
							Tenant: from.Region[10:],
							Name:   from.Name,
						},
						Reason: parts[5],
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
				if len(parts) == 6 {
					message = model.RollbackWasRejected{
						Account: model.Account{
							Tenant: from.Region[10:],
							Name:   from.Name,
						},
						Reason: parts[5],
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

func spawnForwardActor(s *daemon.ActorSystem, name string) (*system.Envelope, error) {
	envelope := system.NewEnvelope(name, model.NewForwardState())

	err := s.RegisterActor(envelope, InitialForward(s))
	if err != nil {
		log.Warnf("%s ~ Spawning Forward Actor Error unable to register", name)
		return nil, err
	}

	log.Debugf("%s ~ Forward Actor Spawned", name)
	return envelope, nil
}

func spawnTransactionActor(s *daemon.ActorSystem, name string) (*system.Envelope, error) {
	envelope := system.NewEnvelope(name, model.NewTransactionState())

	err := s.RegisterActor(envelope, InitialTransaction(s))
	if err != nil {
		log.Warnf("%s ~ Spawning Transaction Actor Error unable to register", name)
		return nil, err
	}

	log.Debugf("%s ~ Transaction Actor Spawned", name)
	return envelope, nil
}

// ProcessLocalMessage processing of local message to this ledger
func ProcessLocalMessage(s *daemon.ActorSystem) system.ProcessLocalMessage {
	return func(message interface{}, to system.Coordinates, from system.Coordinates) {
		if to.Region != "" && to.Region != s.Name {
			log.Warnf("Invalid region received [local %s -> local %s]", from, to)
			return
		}

		ref, err := s.ActorOf(to.Name)
		if err != nil {
			log.Warnf("Actor not found [local %s]", to)
			return
		}
		ref.Tell(message, to, from)
	}
}
