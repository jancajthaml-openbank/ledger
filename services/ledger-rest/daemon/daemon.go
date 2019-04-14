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

package daemon

import (
	"context"
	"time"
)

// Daemon contract for type using support
type Daemon interface {
	WaitReady(deadline time.Duration) error
}

// NewDaemonSupport constructor
func NewDaemonSupport(parentCtx context.Context) Support {
	ctx, cancel := context.WithCancel(parentCtx)
	return Support{
		ctx:        ctx,
		cancel:     cancel,
		exitSignal: make(chan struct{}),
		IsReady:    make(chan interface{}),
		canStart:   make(chan interface{}),
	}
}

// Support provides support for graceful shutdown
type Support struct {
	ctx        context.Context
	cancel     context.CancelFunc
	exitSignal chan struct{}
	IsReady    chan interface{}
	canStart   chan interface{}
}

// GreenLight signals daemon to start work
func (s Support) GreenLight() {
	s.canStart <- nil
}

// MarkDone signals daemon is finished
func (s Support) MarkDone() {
	close(s.exitSignal)
}

// MarkReady signals daemon is ready
func (s Support) MarkReady() {
	s.IsReady <- nil
}

// Done cancel channel
func (s Support) Done() <-chan struct{} {
	return s.ctx.Done()
}

// Stop daemon and wait for graceful shutdown
func (s Support) Stop() {
	s.cancel()
	<-s.exitSignal
}

// Start daemon and wait for it to be ready
func (s Support) Start() {
	s.MarkReady()
	<-s.IsReady
}
