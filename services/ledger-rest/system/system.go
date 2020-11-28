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

package system

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/jancajthaml-openbank/ledger-rest/support/concurrent"
)

// DbusControl is control implementation using dbus
type DbusControl struct {
	Control
	concurrent.DaemonSupport
	underlying *dbus.Conn
}

// NewSystemControl returns new systemctl fascade
func NewSystemControl(ctx context.Context) *DbusControl {
	conn, err := dbus.New()
	if err != nil {
		log.Error().Msgf("Unable to obtain dbus connection because %+v", err)
		return nil
	}
	return &DbusControl{
		DaemonSupport: concurrent.NewDaemonSupport(ctx, "system-control"),
		underlying:    conn,
	}
}

// ListUnits returns list of unit names
func (sys *DbusControl) ListUnits(prefix string) ([]string, error) {
	if sys == nil {
		return nil, fmt.Errorf("cannot call method on nil")
	}

	units, err := sys.underlying.ListUnits()
	if err != nil {
		return nil, err
	}

	fullPrefix := "ledger-" + prefix

	var result = make([]string, 0)
	for _, unit := range units {
		if unit.LoadState == "not-found" || !strings.HasPrefix(unit.Name, fullPrefix) {
			continue
		}
		result = append(result, strings.TrimSuffix(strings.TrimPrefix(unit.Name, fullPrefix), ".service"))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result, nil
}

// GetUnitsProperties return unit properties
func (sys *DbusControl) GetUnitsProperties(prefix string) (map[string]UnitStatus, error) {
	if sys == nil {
		return nil, fmt.Errorf("cannot call method on nil")
	}

	units, err := sys.underlying.ListUnits()
	if err != nil {
		return nil, err
	}

	fullPrefix := "ledger-" + prefix

	var result = make(map[string]UnitStatus)
	for _, unit := range units {
		if !strings.HasPrefix(unit.Name, fullPrefix) {
			continue
		}
		properties, err := sys.underlying.GetUnitProperties(unit.Name)

		if err != nil {
			result[unit.Name] = UnitStatus{
				Status:          unit.SubState,
				StatusChangedAt: 0,
			}
		} else {
			result[unit.Name] = UnitStatus{
				Status:          unit.SubState,
				StatusChangedAt: properties["StateChangeTimestamp"].(uint64),
			}
		}
	}

	return result, nil
}

// DisableUnit disables unit
func (sys *DbusControl) DisableUnit(name string) error {
	if sys == nil {
		return fmt.Errorf("cannot call method on nil")
	}

	ch := make(chan string)

	fullName := "ledger-" + name

	if _, err := sys.underlying.StopUnit(fullName, "replace", ch); err != nil {
		return fmt.Errorf("unable to stop unit %s because %+v", fullName, err)
	}

	select {

	case result := <-ch:
		if result != "done" {
			return fmt.Errorf("unable to stop unit %s", fullName)
		}
		log.Info().Msgf("Stopped unit %s", fullName)
		log.Info().Msgf("Disabling unit %s", fullName)

		if _, err := sys.underlying.DisableUnitFiles([]string{fullName}, false); err != nil {
			return fmt.Errorf("unable to disable unit %s because %+v", fullName, err)
		}

		return nil

	case <-time.After(3 * time.Second):
		return fmt.Errorf("unable to stop unit %s because timeout", fullName)

	}
}

// EnableUnit enables unit
func (sys *DbusControl) EnableUnit(name string) error {
	if sys == nil {
		return fmt.Errorf("cannot call method on nil")
	}

	fullName := "ledger-" + name

	if _, _, err := sys.underlying.EnableUnitFiles([]string{fullName}, false, false); err != nil {
		return fmt.Errorf("unable to enable unit %s because %+v", fullName, err)
	}

	ch := make(chan string)

	if _, err := sys.underlying.StartUnit(fullName, "replace", ch); err != nil {
		return fmt.Errorf("unable to start unit %s because %+v", fullName, err)
	}

	select {

	case result := <-ch:
		if result != "done" {
			return fmt.Errorf("unable to start unit %s", fullName)
		}
		log.Info().Msgf("Started unit %s", fullName)
		return nil

	case <-time.After(3 * time.Second):
		return fmt.Errorf("unable to start unit %s because timeout", fullName)

	}

	return nil
}

// Start handles everything needed to start system-control daemon
func (sys *DbusControl) Start() {
	if sys == nil {
		return
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	sys.MarkReady()

	select {
	case <-sys.CanStart:
		break
	case <-sys.Done():
		sys.MarkDone()
		return
	}

	log.Info().Msg("Start system-control daemon")

	go func() {
		for {
			select {
			case <-sys.Done():
				sys.MarkDone()
				return
			}
		}
	}()

	sys.WaitStop()
	log.Info().Msg("Stop system-control daemon")
}
