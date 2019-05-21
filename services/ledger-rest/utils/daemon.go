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

package utils

import (
	"context"
	"time"
)

// Daemon contract for type using support
type Daemon interface {
	WaitReady(deadline time.Duration) error
}

// Support provides support for graceful shutdown
type DaemonSupport struct {
	ctx        context.Context
	Cancel     context.CancelFunc
	ExitSignal chan struct{}
	IsReady    chan interface{}
	CanStart   chan interface{}
}

// NewDaemonSupport constructor
func NewDaemonSupport(parentCtx context.Context) DaemonSupport {
	ctx, cancel := context.WithCancel(parentCtx)
	return DaemonSupport{
		ctx:        ctx,
		Cancel:     cancel,
		ExitSignal: make(chan struct{}),
		IsReady:    make(chan interface{}),
		CanStart:   make(chan interface{}),
	}
}

// GreenLight signals daemon to start work
func (daemon DaemonSupport) GreenLight() {
	daemon.CanStart <- nil
}

// MarkDone signals daemon is finished
func (daemon DaemonSupport) MarkDone() {
	close(daemon.ExitSignal)
}

// MarkReady signals daemon is ready
func (daemon DaemonSupport) MarkReady() {
	daemon.IsReady <- nil
}

// Done cancel channel
func (daemon DaemonSupport) Done() <-chan struct{} {
	return daemon.ctx.Done()
}

// Stop daemon and wait for graceful shutdown
func (daemon DaemonSupport) Stop() {
	daemon.Cancel()
	<-daemon.ExitSignal
}

// Start daemon and wait for it to be ready
func (daemon DaemonSupport) Start() {
	daemon.MarkReady()
	<-daemon.IsReady
}
