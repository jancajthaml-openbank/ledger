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

package concurrent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Daemon contract for program sub routines
type Daemon interface {
	Start()
	Stop()
	GreenLight()
	WaitStop()
	WaitReady(time.Duration) error
}

// DaemonSupport provides support for graceful shutdown
type DaemonSupport struct {
	Daemon
	name              string
	ctx               context.Context
	cancel            context.CancelFunc
	done              chan interface{}
	doneOnce          sync.Once
	ExitSignal        chan struct{}
	IsReady           chan interface{}
	CanStart          chan interface{}
	closeCanStartOnce sync.Once
}

// NewDaemonSupport constructs new daemon support
func NewDaemonSupport(parentCtx context.Context, name string) DaemonSupport {
	ctx, cancel := context.WithCancel(parentCtx)
	return DaemonSupport{
		name:     name,
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan interface{}),
		IsReady:  make(chan interface{}),
		CanStart: make(chan interface{}),
	}
}

// WaitReady wait for daemon to be ready within given deadline
func (daemon DaemonSupport) WaitReady(deadline time.Duration) error {
	ticker := time.NewTicker(deadline)
	select {
	case <-daemon.IsReady:
		ticker.Stop()
		return nil
	case <-ticker.C:
		return fmt.Errorf("%s-daemon was not ready within %v seconds", daemon.name, deadline)
	}
}

// WaitStop cancels context
func (daemon DaemonSupport) WaitStop() {
	<-daemon.done
}

// GreenLight signals daemon to start work
func (daemon DaemonSupport) GreenLight() {
	daemon.closeCanStartOnce.Do(func() {
		close(daemon.CanStart)
	})
}

// MarkDone signals daemon is finished
func (daemon DaemonSupport) MarkDone() {
	daemon.doneOnce.Do(func() {
		close(daemon.done)
	})
}

// IsCanceled returns if daemon is done
func (daemon DaemonSupport) IsCanceled() bool {
	return daemon.ctx.Err() != nil
}

// MarkReady signals daemon is ready
func (daemon DaemonSupport) MarkReady() {
	daemon.IsReady <- nil
}

// Done cancel channel
func (daemon DaemonSupport) Done() <-chan struct{} {
	return daemon.ctx.Done()
}

// Stop cancels context
func (daemon DaemonSupport) Stop() {
	daemon.cancel()
}
