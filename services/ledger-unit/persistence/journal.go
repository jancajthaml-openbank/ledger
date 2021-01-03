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

package persistence

import (
	"github.com/jancajthaml-openbank/ledger-unit/model"

	localfs "github.com/jancajthaml-openbank/local-fs"
)

// LoadTransaction loads transaction from journal
func LoadTransaction(storage localfs.Storage, id string) (*model.Transaction, error) {
	transactionPath := "transaction/" + id
	data, err := storage.ReadFileFully(transactionPath)
	if err != nil {
		return nil, err
	}
	result := new(model.Transaction)
	result.IDTransaction = id
	result.Deserialize(data)
	return result, nil
}

// LoadTransactionState loads transaction status journal
func LoadTransactionState(storage localfs.Storage, id string) (string, error) {
	transactionPath := "transaction/" + id
	data, err := storage.ReadFileFully(transactionPath)
	if err != nil {
		return "", err
	}
	result := new(model.Transaction)
	result.IDTransaction = id
	result.DeserializeState(data)
	return result.State, nil
}

// CreateTransaction persist transaction entity state to storage
func CreateTransaction(storage localfs.Storage, entity *model.Transaction) error {
	transactionPath := "transaction/" + entity.IDTransaction
	data := entity.Serialize()
	return storage.WriteFileExclusive(transactionPath, data)
}

// UpdateTransaction persist update of transaction to disk
func UpdateTransaction(storage localfs.Storage, entity *model.Transaction) error {
	transactionPath := "transaction/" + entity.IDTransaction
	data := entity.Serialize()
	return storage.WriteFile(transactionPath, data)
}
