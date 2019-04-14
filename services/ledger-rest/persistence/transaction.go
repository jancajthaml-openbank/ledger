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

package persistence

import (
	"github.com/jancajthaml-openbank/ledger-rest/model"
	"github.com/jancajthaml-openbank/ledger-rest/utils"

	localfs "github.com/jancajthaml-openbank/local-fs"
)

func LoadTransactions(storage *localfs.Storage, tenant string) ([]string, error) {
	path := utils.TransactionsPath(tenant)
	ok, err := storage.Exists(path)
	if err != nil || !ok {
		return make([]string, 0), nil
	}
	transactions, err := storage.ListDirectory(path, true)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func LoadTransaction(storage *localfs.Storage, tenant, id string) (*model.Transaction, error) {
	dataPath := utils.TransactionPath(tenant, id)
	data, err := storage.ReadFileFully(dataPath)
	if err != nil {
		return nil, err
	}

	statusPath := utils.TransactionStatePath(tenant, id)
	status, err := storage.ReadFileFully(statusPath)
	if err != nil {
		return nil, err
	}

	result := new(model.Transaction)
	result.Deserialise(data, status)
	return result, nil
}
