// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
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

// CreateTransaction creates new transaction
func CreateTransaction(s *daemon.ActorSystem, tenant string, transaction model.Transaction) (result interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("CreateTransaction recovered in %v", r)
			result = nil
		}
	}()

	ch := make(chan interface{})
	defer close(ch)

	name := "transaction/" + xid.New().String()

	envelope := system.NewEnvelope(name, nil)
	defer s.UnregisterActor(envelope.Name)

	s.RegisterActor(envelope, func(state interface{}, context system.Context) {
		ch <- context.Data
	})

	s.SendRemote("LedgerUnit/"+tenant, CreateTransactionMessage(envelope.Name, name, transaction))

	select {

	case result = <-ch:
		return

	case <-time.After(3 * time.Second):
		log.Warnf("Create transaction %s/%s timeout", tenant, transaction.IDTransaction)
		result = new(model.ReplyTimeout)
		return
	}
}
