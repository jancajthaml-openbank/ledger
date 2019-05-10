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

	localfs "github.com/jancajthaml-openbank/local-fs"
	log "github.com/sirupsen/logrus"

	"github.com/jancajthaml-openbank/ledger-unit/actor"
	"github.com/jancajthaml-openbank/ledger-unit/config"
	"github.com/jancajthaml-openbank/ledger-unit/daemon"
	"github.com/jancajthaml-openbank/ledger-unit/utils"
)

// Application encapsulate initialized application
type Application struct {
	cfg                  config.Configuration
	interrupt            chan os.Signal
	metrics              daemon.Metrics
	actorSystem          daemon.ActorSystem
	transactionFinalizer daemon.TransactionFinalizer
	cancel               context.CancelFunc
}

// Initialize application
func Initialize() Application {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := config.GetConfig()

	utils.SetupLogger(cfg.LogLevel)

	log.Infof(">>> Setup <<<")

	storage := localfs.NewStorage(cfg.RootStorage)

	metrics := daemon.NewMetrics(ctx, cfg)

	actorSystem := daemon.NewActorSystem(ctx, cfg, &metrics, &storage)
	actorSystem.Support.RegisterOnLocalMessage(actor.ProcessLocalMessage(&actorSystem))
	actorSystem.Support.RegisterOnRemoteMessage(actor.ProcessRemoteMessage(&actorSystem))

	transactionFinalizer := daemon.NewTransactionFinalizer(ctx, cfg, &metrics, &storage, actor.ProcessLocalMessage(&actorSystem))

	return Application{
		cfg:                  cfg,
		interrupt:            make(chan os.Signal, 1),
		metrics:              metrics,
		actorSystem:          actorSystem,
		transactionFinalizer: transactionFinalizer,
		cancel:               cancel,
	}
}
