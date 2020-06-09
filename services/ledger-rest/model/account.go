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
	"fmt"
	"strings"

	"github.com/jancajthaml-openbank/ledger-rest/utils"
)

type Account struct {
	Tenant string `json:"tenant"`
	Name   string `json:"name"`
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
