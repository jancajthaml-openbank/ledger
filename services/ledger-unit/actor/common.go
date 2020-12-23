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
	"fmt"

	"github.com/jancajthaml-openbank/ledger-unit/model"

	system "github.com/jancajthaml-openbank/actor-system"
)

func parseTransfer(chunk string) (*model.Transfer, error) {
	start := 0
	end := len(chunk)
	parts := make([]string, 8)
	idx := 0
	i := 0
	for i < end && idx < 8 {
		if chunk[i] == 59 {
			if !(start == i && chunk[start] == 59) {
				parts[idx] = chunk[start:i]
				idx++
			}
			start = i + 1
		}
		i++
	}
	if idx < 8 && chunk[start] != 59 && len(chunk[start:]) > 0 {
		parts[idx] = chunk[start:]
	}

	amount := new(model.Dec)
	if !amount.SetString(parts[5]) {
		return nil, fmt.Errorf("invalid amount %s", parts[5])
	}

	return &model.Transfer{
		IDTransfer: parts[0],
		Credit: model.Account{
			Tenant: parts[1],
			Name:   parts[2],
		},
		Debit: model.Account{
			Tenant: parts[3],
			Name:   parts[4],
		},
		ValueDate: parts[7],
		Amount:    amount,
		Currency:  parts[6],
	}, nil
}

func parseMessage(msg string, from system.Coordinates) (interface{}, error) {
	start := 0
	end := len(msg)
	parts := make([]string, 40)
	idx := 0
	i := 0
	for i < end && idx < 40 {
		if msg[i] == 32 {
			if !(start == i && msg[start] == 32) {
				parts[idx] = msg[start:i]
				idx++
			}
			start = i + 1
		}
		i++
	}
	if idx < 40 && msg[start] != 32 && len(msg[start:]) > 0 {
		parts[idx] = msg[start:]
		idx++
	}

	if i != end {
		return nil, fmt.Errorf("message too large")
	}

	switch parts[0] {

	case ReqCreateTransaction:
		if idx > 2 {
			transaction := model.Transaction{
				IDTransaction: parts[1],
			}
			for _, part := range parts[2:idx] {
				transfer, err := parseTransfer(part)
				if err != nil {
					return nil, fmt.Errorf("invalid transfer in message %s", msg)
				}
				transaction.Transfers = append(transaction.Transfers, *transfer)
			}
			return transaction, nil
		}
		return nil, fmt.Errorf("invalid message %s", msg)

	case FatalError:
		return FatalErrored{
			Account: model.Account{
				Tenant: from.Region[10:],
				Name:   from.Name,
			},
		}, nil

	case PromiseAccepted:
		return PromiseWasAccepted{
			Account: model.Account{
				Tenant: from.Region[10:],
				Name:   from.Name,
			},
		}, nil

	case PromiseRejected:
		if idx == 2 {
			return PromiseWasRejected{
				Account: model.Account{
					Tenant: from.Region[10:],
					Name:   from.Name,
				},
				Reason: parts[1],
			}, nil
		}
		return nil, fmt.Errorf("invalid message %s", msg)

	case CommitAccepted:
		return CommitWasAccepted{
			Account: model.Account{
				Tenant: from.Region[10:],
				Name:   from.Name,
			},
		}, nil

	case CommitRejected:
		if idx == 2 {
			return CommitWasRejected{
				Account: model.Account{
					Tenant: from.Region[10:],
					Name:   from.Name,
				},
				Reason: parts[1],
			}, nil
		}
		return nil, fmt.Errorf("invalid message %s", msg)

	case RollbackAccepted:
		return RollbackWasAccepted{
			Account: model.Account{
				Tenant: from.Region[10:],
				Name:   from.Name,
			},
		}, nil

	case RollbackRejected:
		if idx == 2 {
			return RollbackWasRejected{
				Account: model.Account{
					Tenant: from.Region[10:],
					Name:   from.Name,
				},
				Reason: parts[1],
			}, nil
		}
		return nil, fmt.Errorf("invalid message %s", msg)

	default:
		return nil, fmt.Errorf("invalid message %s", msg)

	}

}

// ProcessMessage processing of remote message to this wall
func ProcessMessage(s *System) system.ProcessMessage {
	return func(msg string, to system.Coordinates, from system.Coordinates) {
		message, err := parseMessage(msg, from)
		if err != nil {
			log.Warn().Msgf("%s [remote %v -> local %v]", err, from, to)
			s.SendMessage(FatalError, from, to)
			return
		}
		var ref *system.Actor
		switch message.(type) {
		case model.Transaction:
			if ref, err = NewTransactionActor(s, to.Name); err != nil {
				log.Warn().Msgf("%s [remote %v -> local %v]", err, from, to)
				s.SendMessage(FatalError, from, to)
				return
			}
		default:
			if ref, err = s.ActorOf(to.Name); err != nil {
				log.Warn().Msgf("Actor not found [remote %v -> local %v]", from, to)
				return
			}
		}
		ref.Tell(message, to, from)
		return
	}
}

// NewTransactionActor creates new transaction actor
func NewTransactionActor(s *System, name string) (*system.Actor, error) {
	envelope := system.NewActor(name, NewTransactionState())
	err := s.RegisterActor(envelope, InitialTransaction(s))
	if err != nil {
		log.Warn().Msgf("Unable to register %s actor", name)
		return nil, err
	}
	log.Debug().Msgf("Actor %s registered", name)
	return envelope, nil
}
