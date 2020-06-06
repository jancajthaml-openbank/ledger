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

package metrics

import (
	"bytes"
	"fmt"
	"github.com/jancajthaml-openbank/ledger-unit/utils"
	"os"
	"strconv"
)

// MarshalJSON serialises Metrics as json bytes
func (metrics *Metrics) MarshalJSON() ([]byte, error) {
	if metrics == nil {
		return nil, fmt.Errorf("cannot marshall nil")
	}

	if metrics.promisedTransactions == nil || metrics.promisedTransfers == nil ||
		metrics.committedTransactions == nil || metrics.committedTransfers == nil ||
		metrics.rollbackedTransactions == nil || metrics.rollbackedTransfers == nil ||
		metrics.forwardedTransactions == nil || metrics.forwardedTransfers == nil {
		return nil, fmt.Errorf("cannot marshall nil references")
	}

	var buffer bytes.Buffer

	buffer.WriteString("{\"promisedTransactions\":")
	buffer.WriteString(strconv.FormatInt(metrics.promisedTransactions.Count(), 10))
	buffer.WriteString(",\"promisedTransfers\":")
	buffer.WriteString(strconv.FormatInt(metrics.promisedTransfers.Count(), 10))
	buffer.WriteString(",\"committedTransactions\":")
	buffer.WriteString(strconv.FormatInt(metrics.committedTransactions.Count(), 10))
	buffer.WriteString(",\"committedTransfers\":")
	buffer.WriteString(strconv.FormatInt(metrics.committedTransfers.Count(), 10))
	buffer.WriteString(",\"rollbackedTransactions\":")
	buffer.WriteString(strconv.FormatInt(metrics.rollbackedTransactions.Count(), 10))
	buffer.WriteString(",\"rollbackedTransfers\":")
	buffer.WriteString(strconv.FormatInt(metrics.rollbackedTransfers.Count(), 10))
	buffer.WriteString(",\"forwardedTransactions\":")
	buffer.WriteString(strconv.FormatInt(metrics.forwardedTransactions.Count(), 10))
	buffer.WriteString(",\"forwardedTransfers\":")
	buffer.WriteString(strconv.FormatInt(metrics.forwardedTransfers.Count(), 10))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// UnmarshalJSON deserializes Metrics from json bytes
func (metrics *Metrics) UnmarshalJSON(data []byte) error {
	if metrics == nil {
		return fmt.Errorf("cannot unmarshall to nil")
	}

	if metrics.promisedTransactions == nil || metrics.promisedTransfers == nil ||
		metrics.committedTransactions == nil || metrics.committedTransfers == nil ||
		metrics.rollbackedTransactions == nil || metrics.rollbackedTransfers == nil ||
		metrics.forwardedTransactions == nil || metrics.forwardedTransfers == nil {
		return fmt.Errorf("cannot unmarshall to nil references")
	}

	aux := &struct {
		PromisedTransactions   int64 `json:"promisedTransactions"`
		PromisedTransfers      int64 `json:"promisedTransfers"`
		CommittedTransactions  int64 `json:"committedTransactions"`
		CommittedTransfers     int64 `json:"committedTransfers"`
		RollbackedTransactions int64 `json:"rollbackedTransactions"`
		RollbackedTransfers    int64 `json:"rollbackedTransfers"`
		ForwardedTransactions  int64 `json:"forwardedTransactions"`
		ForwardedTransfers     int64 `json:"forwardedTransfers"`
	}{}

	if err := utils.JSON.Unmarshal(data, &aux); err != nil {
		return err
	}

	metrics.promisedTransactions.Clear()
	metrics.promisedTransactions.Inc(aux.PromisedTransactions)
	metrics.promisedTransfers.Clear()
	metrics.promisedTransfers.Inc(aux.PromisedTransfers)
	metrics.committedTransactions.Clear()
	metrics.committedTransactions.Inc(aux.CommittedTransactions)
	metrics.committedTransfers.Clear()
	metrics.committedTransfers.Inc(aux.CommittedTransfers)
	metrics.rollbackedTransactions.Clear()
	metrics.rollbackedTransactions.Inc(aux.RollbackedTransactions)
	metrics.rollbackedTransfers.Clear()
	metrics.rollbackedTransfers.Inc(aux.RollbackedTransfers)
	metrics.forwardedTransactions.Clear()
	metrics.forwardedTransactions.Inc(aux.ForwardedTransactions)
	metrics.forwardedTransfers.Clear()
	metrics.forwardedTransfers.Inc(aux.ForwardedTransfers)

	return nil
}

// Persist saved metrics state to storage
func (metrics *Metrics) Persist() error {
	if metrics == nil {
		return fmt.Errorf("cannot persist nil reference")
	}
	data, err := utils.JSON.Marshal(metrics)
	if err != nil {
		return err
	}
	err = metrics.storage.WriteFile("metrics."+metrics.tenant+".json", data)
	if err != nil {
		return err
	}
	err = os.Chmod(metrics.storage.Root+"/metrics."+metrics.tenant+".json", 0644)
	if err != nil {
		return err
	}
	return nil
}

// Hydrate loads metrics state from storage
func (metrics *Metrics) Hydrate() error {
	if metrics == nil {
		return fmt.Errorf("cannot hydrate nil reference")
	}
	data, err := metrics.storage.ReadFileFully("metrics." + metrics.tenant + ".json")
	if err != nil {
		return err
	}
	err = utils.JSON.Unmarshal(data, metrics)
	if err != nil {
		return err
	}
	return nil
}
