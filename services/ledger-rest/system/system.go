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
	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/jancajthaml-openbank/ledger-rest/utils"
	"sort"
	"strings"
	"time"
)

// Control represents systemctl subroutine
type Control struct {
	utils.DaemonSupport
	underlying *dbus.Conn
}

// NewSystemControl returns new systemctl fascade
func NewSystemControl(ctx context.Context) *Control {
	conn, err := dbus.New()
	if err != nil {
		log.Error().Msgf("Unable to obtain dbus connection because %+v", err)
		return nil
	}
	return &Control{
		DaemonSupport: utils.NewDaemonSupport(ctx, "system-control"),
		underlying:    conn,
	}
}

// ListUnits returns list of unit names
func (sys Control) ListUnits(prefix string) ([]string, error) {
	units, err := sys.underlying.ListUnits()
	if err != nil {
		return nil, err
	}

	var result = make([]string, 0)
	for _, unit := range units {
		if unit.LoadState == "not-found" || !strings.HasPrefix(unit.Name, prefix) {
			continue
		}
		result = append(result, strings.TrimSuffix(strings.TrimPrefix(unit.Name, prefix), ".service"))
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result, nil
}

// GetUnitsProperties return unit properties
func (sys Control) GetUnitsProperties(prefix string) (map[string]UnitStatus, error) {
	units, err := sys.underlying.ListUnits()
	if err != nil {
		return nil, err
	}

	var result = make(map[string]UnitStatus)
	for _, unit := range units {
		if !strings.HasPrefix(unit.Name, prefix) {
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
func (sys Control) DisableUnit(name string) error {
	ch := make(chan string)

	if _, err := sys.underlying.StopUnit(name, "replace", ch); err != nil {
		return fmt.Errorf("unable to stop unit %s because %+v", name, err)
	}

	select {

	case result := <-ch:
		if result != "done" {
			return fmt.Errorf("unable to stop unit %s", name)
		}
		log.Info().Msgf("Stopped unit %s", name)
		log.Info().Msgf("Disabling unit %s", name)

		if _, err := sys.underlying.DisableUnitFiles([]string{name}, false); err != nil {
			return fmt.Errorf("unable to disable unit %s because %+v", name, err)
		}

		return nil

	case <-time.After(3 * time.Second):
		return fmt.Errorf("unable to stop unit %s because timeout", name)

	}
}

// EnableUnit enables unit
func (sys Control) EnableUnit(name string) error {
	if _, _, err := sys.underlying.EnableUnitFiles([]string{name}, false, false); err != nil {
		return fmt.Errorf("unable to enable unit %s because %+v", name, err)
	}

	ch := make(chan string)

	if _, err := sys.underlying.StartUnit(name, "replace", ch); err != nil {
		return fmt.Errorf("unable to start unit %s because %+v", name, err)
	}

	select {

	case result := <-ch:
		if result != "done" {
			return fmt.Errorf("unable to start unit %s", name)
		}
		log.Info().Msgf("Started unit %s", name)
		return nil

	case <-time.After(3 * time.Second):
		return fmt.Errorf("unable to start unit %s because timeout", name)

	}

	return nil
}

// Start handles everything needed to start system-control daemon
func (sys Control) Start() {
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
