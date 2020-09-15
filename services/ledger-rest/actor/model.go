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

package actor

// ReplyTimeout message
type ReplyTimeout struct{}

// TransactionCreated message
type TransactionCreated struct{}

// TransactionRace message
type TransactionRace struct{}

// TransactionRefused message
type TransactionRefused struct{}

// TransactionRejected message
type TransactionRejected struct{}

// TransactionDuplicate message
type TransactionDuplicate struct{}

// TransactioMissing message
type TransactioMissing struct{}
