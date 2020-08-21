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

package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/jancajthaml-openbank/ledger-rest/actor"
	"github.com/jancajthaml-openbank/ledger-rest/system"
	"github.com/jancajthaml-openbank/ledger-rest/utils"

	localfs "github.com/jancajthaml-openbank/local-fs"
	"github.com/labstack/echo/v4"
)

// Server is a fascade for http-server following handler api of Gin and
// lifecycle api of http
type Server struct {
	utils.DaemonSupport
	Storage       *localfs.PlaintextStorage
	SystemControl *system.SystemControl
	DiskMonitor   *system.DiskMonitor
	MemoryMonitor *system.MemoryMonitor
	ActorSystem   *actor.ActorSystem
	underlying    *http.Server
	key           []byte
	cert          []byte
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func cloneTLSConfig(cfg *tls.Config) *tls.Config {
	if cfg == nil {
		return cfg
	}
	return &tls.Config{
		Rand:                     cfg.Rand,
		Time:                     cfg.Time,
		Certificates:             cfg.Certificates,
		NameToCertificate:        cfg.NameToCertificate,
		GetCertificate:           cfg.GetCertificate,
		RootCAs:                  cfg.RootCAs,
		NextProtos:               cfg.NextProtos,
		ServerName:               cfg.ServerName,
		ClientAuth:               cfg.ClientAuth,
		ClientCAs:                cfg.ClientCAs,
		InsecureSkipVerify:       cfg.InsecureSkipVerify,
		CipherSuites:             cfg.CipherSuites,
		PreferServerCipherSuites: cfg.PreferServerCipherSuites,
		SessionTicketsDisabled:   cfg.SessionTicketsDisabled,
		SessionTicketKey:         cfg.SessionTicketKey,
		ClientSessionCache:       cfg.ClientSessionCache,
		MinVersion:               cfg.MinVersion,
		MaxVersion:               cfg.MaxVersion,
		CurvePreferences:         cfg.CurvePreferences,
	}
}

// NewServer returns new secure server instance
func NewServer(ctx context.Context, port int, secretsPath string, actorSystem *actor.ActorSystem, systemControl *system.SystemControl, diskMonitor *system.DiskMonitor, memoryMonitor *system.MemoryMonitor, storage *localfs.PlaintextStorage) Server {
	router := echo.New()

	cert, err := ioutil.ReadFile(secretsPath + "/domain.local.crt")
	if err != nil {
		log.Fatalf("unable to load certificate %s/domain.local.crt", secretsPath)
	}

	key, err := ioutil.ReadFile(secretsPath + "/domain.local.key")
	if err != nil {
		log.Fatalf("unable to load certificate %s/domain.local.key", secretsPath)
	}

	router.GET("/health", HealtCheck(memoryMonitor, diskMonitor))
	router.HEAD("/health", HealtCheck(memoryMonitor, diskMonitor))

	router.GET("/tenant", ListTenants(systemControl))
	router.POST("/tenant/:tenant", CreateTenant(systemControl))
	router.DELETE("/tenant/:tenant", DeleteTenant(systemControl))

	router.GET("/transaction/:tenant/:id", GetTransaction(storage))
	router.POST("/transaction/:tenant", CreateTransaction(storage, actorSystem))
	router.GET("/transaction/:tenant", GetTransactions(storage))

	return Server{
		DaemonSupport: utils.NewDaemonSupport(ctx, "http-server"),
		Storage:       storage,
		ActorSystem:   actorSystem,
		SystemControl: systemControl,
		DiskMonitor:   diskMonitor,
		MemoryMonitor: memoryMonitor,
		underlying: &http.Server{
			Addr:         fmt.Sprintf("127.0.0.1:%d", port),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			Handler:      router,
			TLSConfig: &tls.Config{
				MinVersion:               tls.VersionTLS12,
				MaxVersion:               tls.VersionTLS12,
				PreferServerCipherSuites: true,
				CurvePreferences: []tls.CurveID{
					tls.CurveP521,
					tls.CurveP384,
					tls.CurveP256,
				},
				CipherSuites: utils.CipherSuites,
			},
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		},
		key:  key,
		cert: cert,
	}
}

// Start handles everything needed to start http-server daemon
func (server Server) Start() {
	config := cloneTLSConfig(server.underlying.TLSConfig)

	config.Certificates = make([]tls.Certificate, 1)

	cert, err := tls.X509KeyPair(server.cert, server.key)
	if err != nil {
		return
	}

	config.Certificates[0] = cert

	ln, err := net.Listen("tcp", server.underlying.Addr)
	if err != nil {
		return
	}
	defer ln.Close()

	server.MarkReady()

	select {
	case <-server.CanStart:
		break
	case <-server.Done():
		server.MarkDone()
		return
	}

	go func() {
		log.Infof("Start http-server daemon, listening on 127.0.0.1:%d", ln.Addr().(*net.TCPAddr).Port)
		tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, config)
		err := server.underlying.Serve(tlsListener)
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("http-server error %v", err)
			server.Stop()
			return
		}
	}()

	go func() {
		for {
			select {
			case <-server.Done():
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				server.underlying.Shutdown(ctx)
				cancel()
				server.MarkDone()
				return
			}
		}
	}()

	server.WaitStop()
	log.Info("Stop http-server daemon")
}
