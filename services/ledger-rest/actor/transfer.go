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
	"time"

	"github.com/rs/xid"

	system "github.com/jancajthaml-openbank/actor-system"
)

// ForwardTransfer forward existing transfer to different vault
func ForwardTransfer(sys *ActorSystem, tenant string, transaction string, transfer string, forward TransferForward) (result interface{}) {
	sys.Metrics.TimeForwardTransfer(func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("ForwardTransfer recovered in %+v", r)
				result = nil
			}
		}()

		ch := make(chan interface{})
		defer close(ch)

		envelope := system.NewEnvelope("forward/" + xid.New().String(), nil)
		defer sys.UnregisterActor(envelope.Name)

		sys.RegisterActor(envelope, func(state interface{}, context system.Context) {
			ch <- context.Data
		})

		sys.SendMessage(
			ForwardTransferMessage(transaction, transfer, forward),
			system.Coordinates{
				Region: "LedgerUnit/" + tenant,
				Name: envelope.Name,
			},
			system.Coordinates{
				Region: "LedgerRest",
				Name: envelope.Name,
			},
		)

		select {

		case result = <-ch:
			return

		case <-time.After(30 * time.Second):
			log.Warnf("Forward transfer %s/%s/%s timeout", tenant, transaction, transfer)
			result = new(ReplyTimeout)
			return
		}
	})
	return
}
