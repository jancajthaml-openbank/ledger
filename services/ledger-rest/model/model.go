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

package model

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/utils"

	"github.com/rs/xid"
	money "gopkg.in/inf.v0"
)

type ReplyTimeout struct{}

type TransactionCreated struct{}

type TransactionRace struct{}

type TransactionRefused struct{}

type TransactionRejected struct{}

type TransactionDuplicate struct{}

type TransactioMissing struct{}

// Transaction represents egress message of transaction
type Transaction struct {
	IDTransaction string     `json:"id"`
	Status        string     `json:"status,omitempty"`
	Transfers     []Transfer `json:"transfers"`
}

type Transfer struct {
	IDTransfer string     `json:"id"`
	Credit     Account    `json:"credit"`
	Debit      Account    `json:"debit"`
	ValueDate  time.Time  `json:"valueDate"`
	Amount     *money.Dec `json:"amount"`
	Currency   string     `json:"currency"`
}

type TransferForward struct {
	Side          string  `json:"side"`
	TargetAccount Account `json:"target"`
}

type Account struct {
	Tenant string `json:"tenant"`
	Name   string `json:"name"`
}

// UnmarshalJSON is json TransferForward unmarhalling companion
func (entity *TransferForward) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}

	all := struct {
		Side          *string  `json:"side"`
		TargetAccount *Account `json:"target"`
	}{}

	err := utils.JSON.Unmarshal(data, &all)
	if err != nil {
		return err
	}

	if all.Side == nil {
		return fmt.Errorf("required field \"side\" is missing")
	}

	if *all.Side != "credit" && *all.Side != "debit" {
		return fmt.Errorf("invalid field \"side\" value")
	}

	if all.TargetAccount == nil {
		return fmt.Errorf("required field \"target\" is missing")
	}

	entity.Side = *all.Side
	entity.TargetAccount = *all.TargetAccount

	return nil
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

	err := utils.JSON.Unmarshal(data, &all)
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

	err := utils.JSON.Unmarshal(data, &all)
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

// UnmarshalJSON is json Account unmarhalling companion
func (entity *Account) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}

	all := struct {
		Tenant string `json:"tenant"`
		Name   string `json:"name"`
	}{}

	err := utils.JSON.Unmarshal(data, &all)
	if err != nil {
		return err
	}

	if all.Tenant == "" {
		return fmt.Errorf("required field \"tenant\" is missing")
	}
	if all.Name == "" {
		return fmt.Errorf("required field \"name\" is missing")
	}

	entity.Tenant = all.Tenant
	entity.Name = strings.Replace(all.Name, " ", "_", -1)
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

// Deserialise transaction from binary data
func (entity *Transaction) Deserialise(data []byte, status []byte) {
	lines := strings.Split(string(data), "\n")
	//entity.IDTransaction = lines[0]
	entity.Transfers = make([]Transfer, len(lines)-1)

	parts := strings.Split(string(status), " ")
	entity.Status = parts[0]

	for i := range entity.Transfers {
		transfer := strings.SplitN(lines[i], " ", 8)

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
