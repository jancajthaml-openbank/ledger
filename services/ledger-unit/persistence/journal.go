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

package persistence

import (
	"strings"

	"github.com/jancajthaml-openbank/ledger-unit/model"
	"github.com/jancajthaml-openbank/ledger-unit/utils"

	localfs "github.com/jancajthaml-openbank/local-fs"
)

// LoadTransaction loads transaction from journal
func LoadTransaction(storage *localfs.PlaintextStorage, id string) (*model.Transaction, error) {
	transactionPath := utils.TransactionPath(id)

	data, err := storage.ReadFileFully(transactionPath)
	if err != nil {
		return nil, err
	}

	result := new(model.Transaction)
	result.IDTransaction = id
	result.Deserialise(data)
	return result, nil
}

// CreateTransaction persist transaction entity state to storage
func CreateTransaction(storage *localfs.PlaintextStorage) *model.Transaction {
	entity := new(model.Transaction)
	entity.State = model.StatusNew
	return PersistTransaction(storage, entity)
}

// PersistTransaction persist transaction to disk
func PersistTransaction(storage *localfs.PlaintextStorage, entity *model.Transaction) *model.Transaction {
	//created := now()
	// FIXME do not store transaction like this :/ or do so for integrity?

	transactionPath := utils.TransactionPath(entity.IDTransaction)

	data := entity.Serialise()
	if storage.WriteFileExclusive(transactionPath, data) != nil {
		return nil
	}

	return entity
}

// UpdateTransaction persist update of transaction to disk
func UpdateTransaction(storage *localfs.PlaintextStorage, entity *model.Transaction) *model.Transaction {
	//created := now()
	// FIXME do not store transaction like this :/ or do so for integrity?

	transactionPath := utils.TransactionPath(entity.IDTransaction)

	data := entity.Serialise()
	if storage.WriteFile(transactionPath, data) != nil {
		return nil
	}

	return entity
}

// IsTransferForwardedCredit returns true if transaction's credit side was forwarded
func IsTransferForwardedCredit(storage *localfs.PlaintextStorage, idTransaction, idTransfer string) (bool, error) {
	fullPath := utils.TransactionForwardPath(idTransaction)
	ok, err := storage.Exists(fullPath)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	data, err := storage.ReadFileFully(fullPath)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Split(line, " ")

		if idTransfer == parts[0] && parts[1] == "credit" {
			return true, nil
		}
	}
	return false, nil
}

// IsTransferForwardedDebit returns true if transaction's debit side was forwarded
func IsTransferForwardedDebit(storage *localfs.PlaintextStorage, idTransaction, idTransfer string) (bool, error) {
	fullPath := utils.TransactionForwardPath(idTransaction)
	ok, err := storage.Exists(fullPath)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	data, err := storage.ReadFileFully(fullPath)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.Split(line, " ")

		if idTransfer == parts[0] && parts[1] == "debit" {
			return true, nil
		}
	}
	return false, nil
}

// AcceptForwardCredit accepts transaction credit forward request
func AcceptForwardCredit(storage *localfs.PlaintextStorage, targetTenant, targetTransaction, targetTransfer, originTransaction, originTransfer string) error {
	fullPath := utils.TransactionForwardPath(originTransaction)
	return storage.AppendFile(fullPath, []byte(originTransfer+" credit "+targetTenant+" "+targetTransaction+" "+targetTransfer))
}

// AcceptForwardDebit accepts transaction debit forward request
func AcceptForwardDebit(storage *localfs.PlaintextStorage, targetTenant, targetTransaction, targetTransfer, originTransaction, originTransfer string) error {
	fullPath := utils.TransactionForwardPath(originTransaction)
	return storage.AppendFile(fullPath, []byte(originTransfer+" debit "+targetTenant+" "+targetTransaction+" "+targetTransfer))
}
