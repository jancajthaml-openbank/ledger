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

import (
	"github.com/jancajthaml-openbank/ledger-unit/model"
)

// FatalErrored is inbound message that there was a fatal error
type FatalErrored struct {
	Account model.Account
}

// PromiseWasAccepted is inbound message that promise was accepted
type PromiseWasAccepted struct {
	Account model.Account
}

// PromiseWasRejected is inbound message that promise was rejected
type PromiseWasRejected struct {
	Account model.Account
	Reason  string
}

// CommitWasAccepted is inbound message that commit was rejected
type CommitWasAccepted struct {
	Account model.Account
}

// CommitWasRejected is inbound message that commit was rejected
type CommitWasRejected struct {
	Account model.Account
	Reason  string
}

// RollbackWasAccepted is inbound message that rollback was rejected
type RollbackWasAccepted struct {
	Account model.Account
}

// RollbackWasRejected is inbound message that rollback was rejected
type RollbackWasRejected struct {
	Account model.Account
	Reason  string
}
