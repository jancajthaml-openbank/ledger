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

	"github.com/jancajthaml-openbank/ledger-rest/daemon"
	"github.com/jancajthaml-openbank/ledger-rest/model"

	"github.com/rs/xid"

	system "github.com/jancajthaml-openbank/actor-system"
	log "github.com/sirupsen/logrus"
)

// ForwardTransfer forward existing transfer to different vault
func ForwardTransfer(s *daemon.ActorSystem, tenant, transaction, transfer string, forward model.TransferForward) (result interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("ForwardTransfer recovered in %v", r)
			result = nil
		}
	}()

	ch := make(chan interface{})
	defer close(ch)

	name := "forward/" + xid.New().String()

	envelope := system.NewEnvelope(name, nil)
	defer s.UnregisterActor(envelope.Name)

	s.RegisterActor(envelope, func(state interface{}, context system.Context) {
		ch <- context.Data
	})

	s.SendRemote("LedgerUnit/"+tenant, ForwardTransferMessage(envelope.Name, name, transaction, transfer, forward))

	select {

	case result = <-ch:
		return

	case <-time.After(3 * time.Second):
		log.Warnf("Forward transfer %s/%s/%s timeout", tenant, transaction, transfer)
		result = new(model.ReplyTimeout)
		return
	}
}
