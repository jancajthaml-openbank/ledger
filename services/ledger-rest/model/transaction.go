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

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rs/xid"
	money "gopkg.in/inf.v0"
)

// Transaction represents transaction
type Transaction struct {
	IDTransaction string     `json:"id"`
	Status        string     `json:"status,omitempty"`
	Transfers     []Transfer `json:"transfers"`
}

// Transfer represents transfer
type Transfer struct {
	IDTransfer string     `json:"id"`
	Credit     Account    `json:"credit"`
	Debit      Account    `json:"debit"`
	ValueDate  time.Time  `json:"valueDate"`
	Amount     *money.Dec `json:"amount"`
	Currency   string     `json:"currency"`
}

// UnmarshalJSON is json Transaction unmarhalling companion
func (entity *Transaction) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}

	all := struct {
		IDTransaction *string    `json:"id"`
		Transfers     []Transfer `json:"transfers"`
	}{}

	err := json.Unmarshal(data, &all)
	if err != nil {
		return err
	}

	if all.IDTransaction != nil {
		entity.IDTransaction = *all.IDTransaction
	} else {
		entity.IDTransaction = xid.New().String()
	}

	entity.Transfers = all.Transfers

	return nil
}

// UnmarshalJSON is json Transfer unmarhalling companion
func (entity *Transfer) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}

	all := struct {
		ID        *string  `json:"id"`
		Credit    *Account `json:"credit"`
		Debit     *Account `json:"debit"`
		ValueDate *string  `json:"valueDate"`
		Amount    *string  `json:"amount"`
		Currency  *string  `json:"currency"`
	}{}

	err := json.Unmarshal(data, &all)
	if err != nil {
		return err
	}
	if all.ID == nil {
		entity.IDTransfer = xid.New().String()
	} else {
		entity.IDTransfer = *all.ID
	}
	if all.Credit == nil {
		return fmt.Errorf("required field \"credit\" is missing")
	}
	if all.Debit == nil {
		return fmt.Errorf("required field \"debit\" is missing")
	}
	if all.Amount == nil {
		return fmt.Errorf("required field \"amount\" is missing")
	}
	if all.Currency == nil {
		return fmt.Errorf("required field \"currency\" is missing")
	}

	amount, ok := new(money.Dec).SetString(*all.Amount)
	if !ok {
		return fmt.Errorf("invalid amount")
	}

	entity.Credit = *all.Credit
	entity.Debit = *all.Debit
	entity.Amount = amount
	entity.Currency = *all.Currency

	if all.ValueDate == nil {
		entity.ValueDate = time.Now()
		return nil
	}

	t1, err := time.Parse(time.RFC3339, *all.ValueDate)
	if err != nil {
		entity.ValueDate = time.Now()
		return nil
	}

	entity.ValueDate = t1.UTC()
	return nil
}

// MarshalJSON is json Transfer marhalling companion
func (entity Transfer) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer

	buffer.WriteString("{\"id\":\"")
	buffer.WriteString(entity.IDTransfer)
	buffer.WriteString("\",\"credit\":{\"tenant\":\"")
	buffer.WriteString(entity.Credit.Tenant)
	buffer.WriteString("\",\"name\":\"")
	buffer.WriteString(entity.Credit.Name)
	buffer.WriteString("\"},\"debit\":{\"tenant\":\"")
	buffer.WriteString(entity.Debit.Tenant)
	buffer.WriteString("\",\"name\":\"")
	buffer.WriteString(entity.Debit.Name)
	buffer.WriteString("\"},\"valueDate\":\"")
	buffer.WriteString(entity.ValueDate.Format(time.RFC3339))
	buffer.WriteString("\",\"amount\":\"")
	buffer.WriteString(entity.Amount.String())
	buffer.WriteString("\",\"currency\":\"")
	buffer.WriteString(entity.Currency)
	buffer.WriteString("\"}")

	return buffer.Bytes(), nil
}

// Deserialize transaction from binary data
func (entity *Transaction) Deserialize(data []byte) {
	lines := strings.Split(string(data), "\n")
	entity.Status = lines[0]
	entity.Transfers = make([]Transfer, len(lines)-2)

	for i := range entity.Transfers {
		transfer := strings.SplitN(lines[i+1], " ", 8)

		valueDate, _ := time.Parse(time.RFC3339, transfer[5])
		amount, _ := new(money.Dec).SetString(transfer[6])

		entity.Transfers[i] = Transfer{
			IDTransfer: transfer[0],
			Credit: Account{
				Tenant: transfer[1],
				Name:   transfer[2],
			},
			Debit: Account{
				Tenant: transfer[3],
				Name:   transfer[4],
			},
			ValueDate: valueDate,
			Amount:    amount,
			Currency:  transfer[7],
		}
	}

	return
}
