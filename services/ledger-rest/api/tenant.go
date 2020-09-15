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

package api

import (
	"fmt"
	"net/http"

	"github.com/jancajthaml-openbank/ledger-rest/system"

	"github.com/labstack/echo/v4"
)

// CreateTenant enables ledger-unit@{tenant}
func CreateTenant(systemctl *system.Control) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		tenant := c.Param("tenant")
		if tenant == "" {
			return fmt.Errorf("missing tenant")
		}

		err := systemctl.EnableUnit("ledger-unit@" + tenant + ".service")
		if err != nil {
			return err
		}

		c.Response().WriteHeader(http.StatusOK)

		return nil
	}
}

// DeleteTenant disables ledger-unit@{tenant}
func DeleteTenant(systemctl *system.Control) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		tenant := c.Param("tenant")
		if tenant == "" {
			return fmt.Errorf("missing tenant")
		}

		err := systemctl.DisableUnit("ledger-unit@" + tenant + ".service")
		if err != nil {
			return err
		}

		c.Response().WriteHeader(http.StatusOK)

		return nil
	}
}

// ListTenants lists ledger-unit@
func ListTenants(systemctl *system.Control) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		units, err := systemctl.ListUnits("ledger-unit@")
		if err != nil {
			return err
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)

		for idx, unit := range units {
			if idx == len(units)-1 {
				c.Response().Write([]byte(unit))
			} else {
				c.Response().Write([]byte(unit + "\n"))
			}
			c.Response().Flush()
		}

		return nil
	}
}
