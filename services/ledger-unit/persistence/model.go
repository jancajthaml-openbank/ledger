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

package persistence

const (
	// EventPromise represents promise prefix
	EventPromise = "0"
	// EventCommit represents commit prefix
	EventCommit = "1"
	// EventRollback represents rollback prefix
	EventRollback = "2"

	// StatusNew represents NEW transaction
	StatusNew = "new"
	// StatusAccepted represents ACCEPTED transaction
	StatusAccepted = "accepted"
	// StatusRejected represents REJECTED transaction
	StatusRejected = "rejected"
	// StatusCommitted represents COMMITTED transaction
	StatusCommitted = "committed"
	// StatusRollbacked represents ROLLBACKED transaction
	StatusRollbacked = "rollbacked"
)
