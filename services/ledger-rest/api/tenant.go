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

package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jancajthaml-openbank/ledger-rest/daemon"
	"github.com/jancajthaml-openbank/ledger-rest/utils"
	"github.com/labstack/gommon/log"
)

// TenantPartial returns http handler for single tenant
func TenantPartial(system *daemon.SystemControl) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		tenant := vars["tenant"]

		if tenant == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(emptyJSONObject)
			return
		}

		switch r.Method {

		case "POST":
			EnableUnit(system, tenant, w, r)
			return

		case "DELETE":
			DisableUnit(system, tenant, w, r)
			return

		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write(emptyJSONObject)
			return

		}
	}
}

// TenantsPartial returns http handler for tenants
func TenantsPartial(system *daemon.SystemControl) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		units, err := system.ListUnits("ledger-unit@")
		if err != nil {
			log.Errorf("Error when listing units, %+v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(emptyJSONObject)
			return
		}

		resp, err := utils.JSON.Marshal(units)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(emptyJSONArray)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
		return
	}
}

// EnableUnit enables tenant unit
func EnableUnit(system *daemon.SystemControl, tenant string, w http.ResponseWriter, r *http.Request) {
	err := system.EnableUnit("ledger-unit@" + tenant + ".service")
	if err != nil {
		log.Errorf("Error when enabling unit, %+v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONObject)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(emptyJSONObject)
	return
}

// DisableUnit disables tenant unit
func DisableUnit(system *daemon.SystemControl, tenant string, w http.ResponseWriter, r *http.Request) {
	err := system.DisableUnit("ledger-unit@" + tenant + ".service")
	if err != nil {
		log.Errorf("Error when disabling unit, %+v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONObject)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(emptyJSONObject)
	return
}
