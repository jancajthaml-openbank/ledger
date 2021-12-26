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

package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Side represents side of transactions
type Side struct {
	Tenant    string `json:"tenant"`
	Account   string `json:"account"`
	Amount    string `json:"amount"`
	Currency  string `json:"currency"`
}

// UnmarshalJSON is json Side unmarhalling companion
func (entity *Side) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshal to nil pointer")
	}
	all := struct {
		Tenant   string `json:"tenant"`
		Account  string `json:"account"`
		Amount   string  `json:"amount"`
		Currency string  `json:"currency"`
	}{}
	err := json.Unmarshal(data, &all)
	if err != nil {
		return err
	}
	if all.Tenant == "" {
		return fmt.Errorf("required field \"tenant\" is missing")
	}
	if all.Account == "" {
		return fmt.Errorf("required field \"account\" is missing")
	}
	if all.Amount == "" {
		return fmt.Errorf("required field \"amount\" is missing")
	}
	if all.Currency == "" {
		return fmt.Errorf("required field \"currency\" is missing")
	}
	_, err = strconv.ParseFloat(all.Amount, 64)
	if err != nil {
		return fmt.Errorf("invalid amount")
	}
	entity.Currency = all.Currency
	entity.Amount = all.Amount
	entity.Tenant = all.Tenant
	entity.Account = strings.Replace(all.Account, " ", "_", -1)
	return nil
}
