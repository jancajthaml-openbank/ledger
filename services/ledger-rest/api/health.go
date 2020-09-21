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
	"encoding/json"
	"github.com/jancajthaml-openbank/ledger-rest/system"
	"github.com/labstack/echo/v4"
	"net/http"
)

// HealtCheck returns 200 OK if service is healthy, 503 otherwise
func HealtCheck(memoryMonitor *system.MemoryMonitor, diskMonitor *system.DiskMonitor) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)

		status := system.Status{
			Memory: system.MemoryStatus{
				Free:      memoryMonitor.GetFreeMemory(),
				Used:      memoryMonitor.GetUsedMemory(),
				IsHealthy: memoryMonitor.IsHealthy(),
			},
			Storage: system.StorageStatus{
				Free:      diskMonitor.GetFreeDiskSpace(),
				Used:      diskMonitor.GetUsedDiskSpace(),
				IsHealthy: diskMonitor.IsHealthy(),
			},
		}

		if !status.Storage.IsHealthy || !status.Memory.IsHealthy {
			c.Response().WriteHeader(http.StatusServiceUnavailable)
		} else {
			c.Response().WriteHeader(http.StatusOK)
		}

		chunk, err := json.Marshal(status)
		if err != nil {
			return err
		}
		c.Response().Write(chunk)
		c.Response().Flush()
		return nil
	}
}

// HealtCheckPing returns 200 OK if service is healthy, 503 otherwise
func HealtCheckPing(memoryMonitor *system.MemoryMonitor, diskMonitor *system.DiskMonitor) func(c echo.Context) error {
	return func(c echo.Context) error {
		if !memoryMonitor.IsHealthy() || !diskMonitor.IsHealthy() {
			c.Response().WriteHeader(http.StatusServiceUnavailable)
		} else {
			c.Response().WriteHeader(http.StatusOK)
		}
		return nil
	}
}
