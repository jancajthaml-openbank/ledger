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
	"github.com/jancajthaml-openbank/ledger-rest/metrics"
	"github.com/jancajthaml-openbank/ledger-rest/systemd"
	"github.com/jancajthaml-openbank/ledger-rest/utils"

	"github.com/gorilla/mux"
	localfs "github.com/jancajthaml-openbank/local-fs"
	log "github.com/sirupsen/logrus"
)

// Server is a fascade for http-server following handler api of Gin and lifecycle
// api of http
type Server struct {
	utils.DaemonSupport
	Storage       *localfs.Storage
	SystemControl *systemd.SystemControl
	ActorSystem   *actor.ActorSystem
	Metrics       *metrics.Metrics
	underlying    *http.Server
	router        *mux.Router
	key           []byte
	cert          []byte
}

type Endpoint func(*Server) func(http.ResponseWriter, *http.Request)

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func cloneTLSConfig(cfg *tls.Config) *tls.Config {
	if cfg == nil {
		return &tls.Config{}
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
func NewServer(ctx context.Context, port int, secretsPath string, actorSystem *actor.ActorSystem, systemControl *systemd.SystemControl, metrics *metrics.Metrics, storage *localfs.Storage) Server {
	router := mux.NewRouter()

	cert, err := ioutil.ReadFile(secretsPath + "/domain.local.crt")
	if err != nil {
		log.Fatalf("unable to load certificate %s/domain.local.crt", secretsPath)
	}

	key, err := ioutil.ReadFile(secretsPath + "/domain.local.key")
	if err != nil {
		log.Fatalf("unable to load certificate %s/domain.local.key", secretsPath)
	}

	result := Server{
		DaemonSupport: utils.NewDaemonSupport(ctx),
		Storage:       storage,
		Metrics:       metrics,
		ActorSystem:   actorSystem,
		router:        router,
		SystemControl: systemControl,
		underlying: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
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

	result.HandleFunc("/health", HealtCheck, "GET", "HEAD")
	result.HandleFunc("/tenant/{tenant}", TenantPartial, "POST", "DELETE")
	result.HandleFunc("/tenant", TenantsPartial, "GET")
	result.HandleFunc("/transaction/{tenant}/{transaction}", TransactionPartial, "GET")
	result.HandleFunc("/transaction/{tenant}/{transaction}/{transfer}", TransferPartial, "PATCH")
	result.HandleFunc("/transaction/{tenant}", TransactionsPartial, "POST", "GET")

	return result
}

// HandleFunc registers route
func (server Server) HandleFunc(path string, handle Endpoint, methods ...string) *mux.Route {
	log.Debugf("HTTP route %+v %+v registered", methods, path)
	return server.router.HandleFunc(path, handle(&server)).Methods(methods...)
}

// WaitReady wait for server to be ready
func (server Server) WaitReady(deadline time.Duration) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("unknown panic")
			}
		}
	}()

	ticker := time.NewTicker(deadline)
	select {
	case <-server.IsReady:
		ticker.Stop()
		err = nil
		return
	case <-ticker.C:
		err = fmt.Errorf("daemon was not ready within %v seconds", deadline)
		return
	}
}

// Start handles everything needed to start http-server daemon
func (server Server) Start() {
	defer server.MarkDone()

	config := cloneTLSConfig(server.underlying.TLSConfig)

	config.Certificates = make([]tls.Certificate, 1)

	cert, err := tls.X509KeyPair(server.cert, server.key)
	if err != nil {
		server.MarkReady()
		return
	}

	config.Certificates[0] = cert

	ln, err := net.Listen("tcp", server.underlying.Addr)
	if err != nil {
		return
	}
	defer ln.Close()

	go func() {
		tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, config)
		err := server.underlying.Serve(tlsListener)
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("Server error %v", err)
		}
		log.Info("Stop http-server daemon")
	}()

	server.MarkReady()

	select {
	case <-server.CanStart:
		break
	case <-server.Done():
		return
	}

	log.Infof("Start http-server daemon, listening on :%d", ln.Addr().(*net.TCPAddr).Port)

	<-server.ExitSignal
}

// Stop tries to shut down http-server daemon gracefully within 5 seconds
func (server Server) Stop() {
	log.Info("Stopping http-server daemon")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.underlying.Shutdown(ctx)
	server.Cancel()
	return
}
