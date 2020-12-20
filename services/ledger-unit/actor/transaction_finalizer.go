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
	"time"

	"github.com/jancajthaml-openbank/ledger-unit/model"
	"github.com/jancajthaml-openbank/ledger-unit/persistence"

	localfs "github.com/jancajthaml-openbank/local-fs"
)

// TransactionFinalizer represents journal saturation update subroutine
type TransactionFinalizer struct {
	callback func(transaction model.Transaction)
	storage  localfs.Storage
}

// NewTransactionFinalizer returns snapshot updater fascade
func NewTransactionFinalizer(rootStorage string, callback func(transaction model.Transaction)) *TransactionFinalizer {
	storage, err := localfs.NewPlaintextStorage(rootStorage)
	if err != nil {
		log.Error().Msgf("Failed to ensure storage %+v", err)
		return nil
	}
	return &TransactionFinalizer{
		callback: callback,
		storage:  storage,
	}
}

func (scan *TransactionFinalizer) getTransactions() []string {
	if scan == nil {
		return nil
	}
	result, err := scan.storage.ListDirectory("transaction", true)
	if err != nil {
		return nil
	}
	return result
}

func (scan *TransactionFinalizer) finalizeStaleTransactions() {
	if scan == nil {
		return
	}
	log.Info().Msg("Performing stale transactions scan")
	transactions := scan.getTransactions()
	for _, transaction := range transactions {
		instance := scan.getTransaction(transaction)
		if instance == nil {
			continue
		}
		log.Info().Msgf("Transaction %s in state %s needs completion", transaction, instance.State)
		scan.callback(*instance)
	}
}

func (scan *TransactionFinalizer) getTransaction(id string) *model.Transaction {
	if scan == nil {
		return nil
	}
	modTime, err := scan.storage.LastModification("transaction/" + id)
	if err != nil {
		return nil
	}
	if time.Now().Sub(modTime).Seconds() < 120 { // FIXME configurable
		return nil
	}
	state, err := persistence.LoadTransactionState(scan.storage, id)
	if err != nil {
		return nil
	}
	if state == persistence.StatusCommitted || state == persistence.StatusRollbacked {
		return nil
	}
	transaction, err := persistence.LoadTransaction(scan.storage, id)
	if err != nil {
		return nil
	}
	return transaction
}

// Setup does nothing
func (scan *TransactionFinalizer) Setup() error {
	return nil
}

// Work finalizes stale transactions
func (scan *TransactionFinalizer) Work() {
	if scan == nil {
		return
	}
	scan.finalizeStaleTransactions()
}

// Cancel does nothing
func (scan *TransactionFinalizer) Cancel() {
	if scan == nil {
		return
	}
}

// Done always returns done
func (scan *TransactionFinalizer) Done() <-chan interface{} {
	done := make(chan interface{})
	close(done)
	return done
}
