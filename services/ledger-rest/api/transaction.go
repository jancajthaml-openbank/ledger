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
	"io/ioutil"
	"net/http"

	"github.com/jancajthaml-openbank/ledger-rest/actor"
	"github.com/jancajthaml-openbank/ledger-rest/model"
	"github.com/jancajthaml-openbank/ledger-rest/persistence"
	"github.com/jancajthaml-openbank/ledger-rest/utils"

	"github.com/gorilla/mux"
)

// TransactionPartial returns http handler for single transaction
func TransactionPartial(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		tenant := vars["tenant"]
		transaction := vars["transaction"]

		if tenant == "" || transaction == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(emptyJSONObject)
			return
		}

		switch r.Method {

		case "GET":
			GetTransaction(server, tenant, transaction, w, r)
			return

		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write(emptyJSONObject)
			return

		}
	}
}

// TransactionsPartial returns http handler for transactions
func TransactionsPartial(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		tenant := vars["tenant"]

		if tenant == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write(emptyJSONArray)
			return
		}

		switch r.Method {

		case "GET":
			GetTransactions(server, tenant, w, r)
			return

		case "POST":
			CreateTransaction(server, tenant, w, r)
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
func CreateTransaction(server *Server, tenant string, w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONObject)
		return
	}

	var req = new(model.Transaction)
	err = utils.JSON.Unmarshal(data, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(emptyJSONObject)
		return
	}

	switch actor.CreateTransaction(server.ActorSystem, tenant, *req).(type) {

	case *actor.TransactionCreated:
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

	case *actor.TransactionRejected:
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

	case *actor.TransactionRefused:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusExpectationFailed)
		w.Write(emptyJSONObject)
		return

	case *actor.TransactionDuplicate:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		w.Write(emptyJSONObject)
		return

	case *actor.TransactionRace, *actor.ReplyTimeout:
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

// GetTransactions returns list of existing transactions
func GetTransactions(server *Server, tenant string, w http.ResponseWriter, r *http.Request) {
	transactions, err := persistence.LoadTransactionsIDS(server.Storage, tenant)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONArray)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp, err := utils.JSON.Marshal(transactions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONArray)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
	return
}

// GetTransaction returns single existing transactions
func GetTransaction(server *Server, tenant string, transaction string, w http.ResponseWriter, r *http.Request) {
	transactions, err := persistence.LoadTransaction(server.Storage, tenant, transaction)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONArray)
		return
	}

	if transactions == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(emptyJSONObject)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp, err := utils.JSON.Marshal(transactions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(emptyJSONArray)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
	return
}
