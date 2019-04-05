// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"io/ioutil"
	"net/http"

	"github.com/jancajthaml-openbank/ledger-rest/actor"
	"github.com/jancajthaml-openbank/ledger-rest/daemon"
	"github.com/jancajthaml-openbank/ledger-rest/model"
	"github.com/jancajthaml-openbank/ledger-rest/utils"

	"github.com/gorilla/mux"
)

// TransactionPartial returns http handler for single transfer
func TransferPartial(metrics *daemon.Metrics, system *daemon.ActorSystem) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)

		tenant := vars["tenant"]
		transaction := vars["transaction"]
		transfer := vars["transfer"]

		if tenant == "" || transaction == "" || transfer == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(emptyJSONArray)
			return
		}

		switch r.Method {

		case "PATCH":
			metrics.TimeForwardTransfer(func() {
				ForwardTransfer(system, tenant, transaction, transfer, w, r)
			})
			return

		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write(emptyJSONObject)
			return

		}
	}
}

// CreateTransaction creates new transaction
func ForwardTransfer(system *daemon.ActorSystem, tenant, transaction, transfer string, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONObject)
		return
	}

	var req = new(model.TransferForward)
	err = utils.JSON.Unmarshal(data, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(emptyJSONObject)
		return
	}

	switch actor.ForwardTransfer(system, tenant, transaction, transfer, *req).(type) {

	case *model.TransactioMissing:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(emptyJSONArray)
		return

	case *model.TransactionCreated:
		resp, err := utils.JSON.Marshal(req)
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

	case *model.TransactionRejected:
		resp, err := utils.JSON.Marshal(req)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(emptyJSONArray)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(resp)
		return

	case *model.TransactionRefused:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusExpectationFailed)
		w.Write(emptyJSONObject)
		return

	case *model.TransactionDuplicate:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(emptyJSONObject)
		return

	case *model.TransactionRace, *model.ReplyTimeout:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write(emptyJSONObject)
		return

	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONObject)
		return

	}
	return
}
