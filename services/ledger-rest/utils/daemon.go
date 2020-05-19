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

package utils

import (
	"context"
	"fmt"
	"time"
)

// Daemon contract for type using support
type Daemon interface {
	Start()
	Stop()
	GreenLight()
	WaitStop()
	WaitReady(time.Duration) error
}

// DaemonSupport provides support for graceful shutdown
type DaemonSupport struct {
	name       string
	ctx        context.Context
	cancel     context.CancelFunc
	done       chan interface{}
	ExitSignal chan struct{}
	IsReady    chan interface{}
	CanStart   chan interface{}
}

// NewDaemonSupport constructor
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
func (daemon DaemonSupport) WaitReady(deadline time.Duration) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("%s-daemon unknown panic", daemon.name)
			}
		}
	}()

	ticker := time.NewTicker(deadline)
	select {
	case <-daemon.IsReady:
		ticker.Stop()
		err = nil
		return
	case <-ticker.C:
		err = fmt.Errorf("%s-daemon was not ready within %v seconds", daemon.name, deadline)
		return
	}
}

// WaitStop cancels context
func (daemon DaemonSupport) WaitStop() {
	<-daemon.done
}

// GreenLight signals daemon to start work
func (daemon DaemonSupport) GreenLight() {
	daemon.CanStart <- nil
}

// MarkDone signals daemon is finished
func (daemon DaemonSupport) MarkDone() {
	close(daemon.done)
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
