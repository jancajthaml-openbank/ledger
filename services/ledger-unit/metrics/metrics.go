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

package metrics

import (
	"sync/atomic"

	"github.com/DataDog/datadog-go/statsd"
)

// Metrics provides helper function for metrics
type Metrics interface {
	TransactionPromised(transfers int)
	TransactionCommitted(transfers int)
	TransactionRollbacked(transfers int)
}

// StatsdMetrics provides metrics helper with statsd client
type StatsdMetrics struct {
	client                 *statsd.Client
	tenant                 string
	promisedTransactions   int64
	promisedTransfers      int64
	committedTransactions  int64
	committedTransfers     int64
	rollbackedTransactions int64
	rollbackedTransfers    int64
}

// NewMetrics returns blank metrics holder
func NewMetrics(tenant string, endpoint string) *StatsdMetrics {
	client, err := statsd.New(endpoint, statsd.WithClientSideAggregation(), statsd.WithoutTelemetry())
	if err != nil {
		log.Error().Msgf("Failed to ensure statsd client %+v", err)
		return nil
	}
	return &StatsdMetrics{
		client:                 client,
		tenant:                 tenant,
		promisedTransactions:   int64(0),
		promisedTransfers:      int64(0),
		committedTransactions:  int64(0),
		committedTransfers:     int64(0),
		rollbackedTransactions: int64(0),
		rollbackedTransfers:    int64(0),
	}
}

// TransactionPromised increments transactions promised by one
func (instance *StatsdMetrics) TransactionPromised(transfers int) {
	if instance == nil {
		return
	}
	atomic.AddInt64(&(instance.promisedTransactions), 1)
	atomic.AddInt64(&(instance.promisedTransfers), int64(transfers))
}

// TransactionCommitted increments transactions committed by one
func (instance *StatsdMetrics) TransactionCommitted(transfers int) {
	if instance == nil {
		return
	}
	atomic.AddInt64(&(instance.committedTransactions), 1)
	atomic.AddInt64(&(instance.committedTransfers), int64(transfers))
}

// TransactionRollbacked increments transactions rollbacked by one
func (instance *StatsdMetrics) TransactionRollbacked(transfers int) {
	if instance == nil {
		return
	}
	atomic.AddInt64(&(instance.rollbackedTransactions), 1)
	atomic.AddInt64(&(instance.rollbackedTransfers), int64(transfers))
}

// Setup does nothing
func (*StatsdMetrics) Setup() error {
	return nil
}

// Done returns always finished
func (*StatsdMetrics) Done() <-chan interface{} {
	done := make(chan interface{})
	close(done)
	return done
}

// Cancel does nothing
func (*StatsdMetrics) Cancel() {
}

// Work represents metrics worker work
func (instance *StatsdMetrics) Work() {
	if instance == nil {
		return
	}

	promisedTransactions := instance.promisedTransactions
	promisedTransfers := instance.promisedTransfers
	committedTransactions := instance.committedTransactions
	committedTransfers := instance.committedTransfers
	rollbackedTransactions := instance.rollbackedTransactions
	rollbackedTransfers := instance.rollbackedTransfers

	atomic.AddInt64(&(instance.promisedTransactions), -promisedTransactions)
	atomic.AddInt64(&(instance.promisedTransfers), -promisedTransfers)
	atomic.AddInt64(&(instance.committedTransactions), -committedTransactions)
	atomic.AddInt64(&(instance.committedTransfers), -committedTransfers)
	atomic.AddInt64(&(instance.rollbackedTransactions), -rollbackedTransactions)
	atomic.AddInt64(&(instance.rollbackedTransfers), -rollbackedTransfers)

	tags := []string{"tenant:" + instance.tenant}

	instance.client.Count("openbank.ledger.transaction.promised", promisedTransactions, tags, 1)
	instance.client.Count("openbank.ledger.transfer.promised", promisedTransfers, tags, 1)
	instance.client.Count("openbank.ledger.transaction.committed", committedTransactions, tags, 1)
	instance.client.Count("openbank.ledger.transfer.committed", committedTransfers, tags, 1)
	instance.client.Count("openbank.ledger.transaction.rollbacked", rollbackedTransactions, tags, 1)
	instance.client.Count("openbank.ledger.transfer.rollbacked", rollbackedTransfers, tags, 1)
}
