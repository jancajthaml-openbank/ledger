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

package boot

import (
	"context"
	"os"

	"github.com/jancajthaml-openbank/ledger-rest/actor"
	"github.com/jancajthaml-openbank/ledger-rest/api"
	"github.com/jancajthaml-openbank/ledger-rest/config"
	"github.com/jancajthaml-openbank/ledger-rest/metrics"
	"github.com/jancajthaml-openbank/ledger-rest/system"
	"github.com/jancajthaml-openbank/ledger-rest/utils"

	localfs "github.com/jancajthaml-openbank/local-fs"
)

// Program encapsulate initialized application
type Program struct {
	interrupt     chan os.Signal
	cfg           config.Configuration
	actorSystem   actor.ActorSystem
	metrics       metrics.Metrics
	rest          api.Server
	systemControl system.SystemControl
	diskMonitor   system.DiskMonitor
	memoryMonitor system.MemoryMonitor
	cancel        context.CancelFunc
}

// Initialize application
func Initialize() Program {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := config.GetConfig()

	utils.SetupLogger(cfg.LogLevel)

	systemControlDaemon := system.NewSystemControl(ctx)
	diskMonitorDaemon := system.NewDiskMonitor(ctx, cfg.MinFreeDiskSpace, cfg.RootStorage)
	memoryMonitorDaemon := system.NewMemoryMonitor(ctx, cfg.MinFreeMemory)

	storage := localfs.NewPlaintextStorage(cfg.RootStorage)
	metricsDaemon := metrics.NewMetrics(ctx, cfg.MetricsOutput, cfg.MetricsRefreshRate)

	actorSystemDaemon := actor.NewActorSystem(ctx, cfg.LakeHostname, &metricsDaemon)
	restDaemon := api.NewServer(ctx, cfg.ServerPort, cfg.SecretsPath, &actorSystemDaemon, &systemControlDaemon, &diskMonitorDaemon, &memoryMonitorDaemon, &storage)

	return Program{
		interrupt:     make(chan os.Signal, 1),
		cfg:           cfg,
		metrics:       metricsDaemon,
		actorSystem:   actorSystemDaemon,
		rest:          restDaemon,
		systemControl: systemControlDaemon,
		diskMonitor:   diskMonitorDaemon,
		memoryMonitor: memoryMonitorDaemon,
		cancel:        cancel,
	}
}
