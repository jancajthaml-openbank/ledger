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

package system

import "github.com/jancajthaml-openbank/ledger-rest/support/logging"

var log = logging.New("system")

// HealthCheck gives insige into system health
type HealthCheck interface {
	IsHealthy() bool
}

// CapacityCheck gives insige into system capacity
type CapacityCheck interface {
	HealthCheck
	GetFree() uint64
	GetUsed() uint64
}
