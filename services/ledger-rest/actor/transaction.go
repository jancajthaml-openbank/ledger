// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	system "github.com/jancajthaml-openbank/actor-system"
	"github.com/jancajthaml-openbank/ledger-rest/model"
	"github.com/rs/xid"
	"time"
)

// CreateTransaction creates new transaction
func CreateTransaction(sys *System, tenant string, transaction model.Transaction) interface{} {

	ch := make(chan interface{})
	defer close(ch)

	envelope := system.NewActor("transaction/"+xid.New().String(), nil)
	defer sys.UnregisterActor(envelope.Name)

	sys.RegisterActor(envelope, func(state interface{}, context system.Context) {
		ch <- context.Data
	})

	sys.SendMessage(
		CreateTransactionMessage(transaction),
		system.Coordinates{
			Region: "LedgerUnit/" + tenant,
			Name:   envelope.Name,
		},
		system.Coordinates{
			Region: "LedgerRest",
			Name:   envelope.Name,
		},
	)

	select {

	case result := <-ch:
		return result

	case <-time.After(20 * time.Second):
		return new(ReplyTimeout)

	}


}
