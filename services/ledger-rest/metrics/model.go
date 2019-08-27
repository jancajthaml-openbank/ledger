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

package metrics

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/utils"
	metrics "github.com/rcrowley/go-metrics"
)

// Metrics represents metrics subroutine
type Metrics struct {
	utils.DaemonSupport
	output                   string
	refreshRate              time.Duration
	createTransactionLatency metrics.Timer
	forwardTransferLatency   metrics.Timer
}

// NewMetrics returns metrics fascade
func NewMetrics(ctx context.Context, output string, refreshRate time.Duration) Metrics {
	return Metrics{
		DaemonSupport:            utils.NewDaemonSupport(ctx),
		output:                   output,
		refreshRate:              refreshRate,
		createTransactionLatency: metrics.NewTimer(),
		forwardTransferLatency:   metrics.NewTimer(),
	}
}

// MarshalJSON serialises Metrics as json bytes
func (metrics *Metrics) MarshalJSON() ([]byte, error) {
	if metrics == nil {
		return nil, fmt.Errorf("cannot marshall nil")
	}

	if metrics.createTransactionLatency == nil || metrics.forwardTransferLatency == nil {
		return nil, fmt.Errorf("cannot marshall nil references")
	}

	var stats = new(runtime.MemStats)
	runtime.ReadMemStats(stats)

	var buffer bytes.Buffer

	buffer.WriteString("{\"createTransactionLatency\":")
	buffer.WriteString(strconv.FormatFloat(metrics.createTransactionLatency.Percentile(0.95), 'f', -1, 64))
	buffer.WriteString(",\"forwardTransferLatency\":")
	buffer.WriteString(strconv.FormatFloat(metrics.forwardTransferLatency.Percentile(0.95), 'f', -1, 64))
	buffer.WriteString(",\"memoryAllocated\":")
	buffer.WriteString(strconv.FormatUint(stats.Sys, 10))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// UnmarshalJSON deserializes Metrics from json bytes
func (metrics *Metrics) UnmarshalJSON(data []byte) error {
	if metrics == nil {
		return fmt.Errorf("cannot unmarshall to nil")
	}

	if metrics.createTransactionLatency == nil || metrics.forwardTransferLatency == nil {
		return fmt.Errorf("cannot unmarshall to nil references")
	}

	aux := &struct {
		GetTransactionLatency    float64 `json:"getTransactionLatency"`
		GetTransactionsLatency   float64 `json:"getTransactionsLatency"`
		CreateTransactionLatency float64 `json:"createTransactionLatency"`
		ForwardTransferLatency   float64 `json:"forwardTransferLatency"`
	}{}

	if err := utils.JSON.Unmarshal(data, &aux); err != nil {
		return err
	}

	metrics.createTransactionLatency.Update(time.Duration(aux.CreateTransactionLatency))
	metrics.forwardTransferLatency.Update(time.Duration(aux.ForwardTransferLatency))

	return nil
}
