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

package metrics

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/jancajthaml-openbank/ledger-unit/utils"
)

// MarshalJSON serialises Metrics as json preserving uint64
func (entity *Metrics) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer

	buffer.WriteString("{\"promisedTransactions\":")
	buffer.WriteString(strconv.FormatInt(entity.promisedTransactions.Count(), 10))
	buffer.WriteString(",\"promisedTransfers\":")
	buffer.WriteString(strconv.FormatInt(entity.promisedTransfers.Count(), 10))
	buffer.WriteString(",\"committedTransactions\":")
	buffer.WriteString(strconv.FormatInt(entity.committedTransactions.Count(), 10))
	buffer.WriteString(",\"committedTransfers\":")
	buffer.WriteString(strconv.FormatInt(entity.committedTransfers.Count(), 10))
	buffer.WriteString(",\"rollbackedTransactions\":")
	buffer.WriteString(strconv.FormatInt(entity.rollbackedTransactions.Count(), 10))
	buffer.WriteString(",\"rollbackedTransfers\":")
	buffer.WriteString(strconv.FormatInt(entity.rollbackedTransfers.Count(), 10))
	buffer.WriteString(",\"forwardedTransactions\":")
	buffer.WriteString(strconv.FormatInt(entity.forwardedTransactions.Count(), 10))
	buffer.WriteString(",\"forwardedTransfers\":")
	buffer.WriteString(strconv.FormatInt(entity.forwardedTransfers.Count(), 10))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// UnmarshalJSON unmarshal json of Metrics entity
func (entity *Metrics) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshall to nil pointer")
	}
	all := struct {
		PromisedTransactions   int64 `json:"promisedTransactions"`
		PromisedTransfers      int64 `json:"promisedTransfers"`
		CommittedTransactions  int64 `json:"committedTransactions"`
		CommittedTransfers     int64 `json:"committedTransfers"`
		RollbackedTransactions int64 `json:"rollbackedTransactions"`
		RollbackedTransfers    int64 `json:"rollbackedTransfers"`
		ForwardedTransactions  int64 `json:"forwardedTransactions"`
		ForwardedTransfers     int64 `json:"forwardedTransfers"`
	}{}
	err := utils.JSON.Unmarshal(data, &all)
	if err != nil {
		return err
	}

	entity.promisedTransactions.Clear()
	entity.promisedTransactions.Inc(all.PromisedTransactions)

	entity.promisedTransfers.Clear()
	entity.promisedTransfers.Inc(all.PromisedTransfers)

	entity.committedTransactions.Clear()
	entity.committedTransactions.Inc(all.CommittedTransactions)

	entity.committedTransfers.Clear()
	entity.committedTransfers.Inc(all.CommittedTransfers)

	entity.rollbackedTransactions.Clear()
	entity.rollbackedTransactions.Inc(all.RollbackedTransactions)

	entity.rollbackedTransfers.Clear()
	entity.rollbackedTransfers.Inc(all.RollbackedTransfers)

	entity.forwardedTransactions.Clear()
	entity.forwardedTransactions.Inc(all.ForwardedTransactions)

	entity.forwardedTransfers.Clear()
	entity.forwardedTransfers.Inc(all.ForwardedTransfers)

	return nil
}

func (metrics *Metrics) Persist() error {
	if metrics == nil {
		return fmt.Errorf("cannot persist nil reference")
	}
	tempFile := metrics.output + "_temp"
	data, err := utils.JSON.Marshal(metrics)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := os.Rename(tempFile, metrics.output); err != nil {
		return err
	}
	return nil
}

func (metrics *Metrics) Hydrate() error {
	if metrics == nil {
		return fmt.Errorf("cannot hydrate nil reference")
	}
	f, err := os.OpenFile(metrics.output, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	buf := make([]byte, fi.Size())
	_, err = f.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}
	err = utils.JSON.Unmarshal(buf, metrics)
	if err != nil {
		return err
	}
	return nil
}
