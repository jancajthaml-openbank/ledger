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

	"github.com/jancajthaml-openbank/ledger-rest/daemon"
	"github.com/jancajthaml-openbank/ledger-rest/model"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
)

var nilCoordinates = system.Coordinates{}

func asEnvelopes(s *daemon.ActorSystem, parts []string) (system.Coordinates, system.Coordinates, string, error) {
	if len(parts) < 4 {
		return nilCoordinates, nilCoordinates, "", fmt.Errorf("invalid message received %+v", parts)
	}

	region, receiver, sender, payload := parts[0], parts[1], parts[2], parts[3]

	from := system.Coordinates{
		Name:   sender,
		Region: region,
	}

	to := system.Coordinates{
		Name:   receiver,
		Region: s.Name,
	}

	return from, to, payload, nil
}

// ProcessRemoteMessage processing of remote message to this wall
func ProcessRemoteMessage(s *daemon.ActorSystem) system.ProcessRemoteMessage {
	return func(parts []string) {
		from, to, payload, err := asEnvelopes(s, parts)
		if err != nil {
			log.Warn(err.Error())
			return
		}

		defer func() {
			if r := recover(); r != nil {
				log.Errorf("procesRemoteMessage recovered in [remote %v -> local %v] : %+v", from, to, r)
			}
		}()

		ref, err := s.ActorOf(to.Name)
		if err != nil {
			// FIXME forward into deadletter receiver and finish whatever has started
			log.Warnf("Deadletter received [remote %v -> local %v] : %+v", from, to, parts[3:])
			return
		}

		var message interface{}

		switch payload {

		case FatalError:
			message = FatalError

		case RespCreateTransaction:
			message = new(model.TransactionCreated)

		case RespTransactionRefused:
			message = new(model.TransactionRefused)

		case RespTransactionRejected:
			message = new(model.TransactionRejected)

		case RespTransactionRace:
			message = new(model.TransactionRace)

		case RespTransactionMissing:
			message = new(model.TransactioMissing)

		case RespTransactionDuplicate:
			message = new(model.TransactionDuplicate)

		default:
			log.Warnf("Deserialization of unsuported message [remote %v -> local %v] : %+v", from, to, parts)
			message = FatalError
		}

		ref.Tell(message, from)
		return
	}
}
