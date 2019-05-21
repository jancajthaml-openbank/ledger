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

	"github.com/jancajthaml-openbank/ledger-rest/model"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
)

func asEnvelopes(s *ActorSystem, msg string) (system.Coordinates, system.Coordinates, []string, error) {
	parts := strings.Split(msg, " ")

	if len(parts) < 5 {
		return system.Coordinates{}, system.Coordinates{}, nil, fmt.Errorf("invalid message received %+v", parts)
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
func ProcessRemoteMessage(s *ActorSystem) system.ProcessRemoteMessage {
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

		ref, err := s.ActorOf(to.Name)
		if err != nil {
			// FIXME forward into deadletter receiver and finish whatever has started
			log.Warnf("Deadletter received [remote %v -> local %v] : %+v", from, to, msg)
			return
		}

		var message interface{}

		switch parts[4] {

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
			log.Warnf("Deserialization of unsuported message [remote %v -> local %v] : %+v", from, to, msg)
			message = FatalError
		}

		ref.Tell(message, to, from)
		return
	}
}
