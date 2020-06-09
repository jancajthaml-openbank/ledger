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
	"net/http"

	"github.com/jancajthaml-openbank/ledger-rest/system"
	"github.com/jancajthaml-openbank/ledger-rest/utils"
)

// HealtCheck returns 200 OK if service is healthy, 503 otherwise
func HealtCheck(server *Server) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		units, err := server.SystemControl.GetUnitsProperties("ledger")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write(emptyJSONObject)
			return
		}

		status := system.SystemStatus{
			Units: units,
			Memory: system.MemoryStatus{
				Free:      server.MemoryMonitor.GetFreeMemory(),
				Used:      server.MemoryMonitor.GetUsedMemory(),
				IsHealthy: server.MemoryMonitor.IsHealthy(),
			},
			Storage: system.StorageStatus{
				Free:      server.DiskMonitor.GetFreeDiskSpace(),
				Used:      server.DiskMonitor.GetUsedDiskSpace(),
				IsHealthy: server.DiskMonitor.IsHealthy(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		resp, err := utils.JSON.Marshal(status)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write(emptyJSONArray)
		} else if !status.Storage.IsHealthy || !status.Memory.IsHealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write(resp)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
		}
	}
}
