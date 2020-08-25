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
	underlying *http.Server
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

// NewServer returns new secure server instance
func NewServer(ctx context.Context, port int, certPath string, keyPath string, actorSystem *actor.ActorSystem, systemControl *system.SystemControl, diskMonitor *system.DiskMonitor, memoryMonitor *system.MemoryMonitor, storage *localfs.PlaintextStorage) Server {
	router := echo.New()

	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Fatalf("Invalid cert %s and key %s", certPath, keyPath)
	}

	router.GET("/health", HealtCheck(memoryMonitor, diskMonitor))
	router.HEAD("/health", HealtCheckPing(memoryMonitor, diskMonitor))

	router.GET("/tenant", ListTenants(systemControl))
	router.POST("/tenant/:tenant", CreateTenant(systemControl))
	router.DELETE("/tenant/:tenant", DeleteTenant(systemControl))

	router.GET("/transaction/:tenant/:id", GetTransaction(storage))
	router.POST("/transaction/:tenant", CreateTransaction(storage, actorSystem))
	router.GET("/transaction/:tenant", GetTransactions(storage))

	return Server{
		DaemonSupport: utils.NewDaemonSupport(ctx, "http-server"),
		underlying: &http.Server{
			Addr:         fmt.Sprintf("127.0.0.1:%d", port),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			Handler:      router,
			TLSConfig: &tls.Config{
				MinVersion:               tls.VersionTLS12,
				MaxVersion:               tls.VersionTLS12,
				PreferServerCipherSuites: true,
				InsecureSkipVerify:       false,
				CurvePreferences: []tls.CurveID{
					tls.CurveP521,
					tls.CurveP384,
					tls.CurveP256,
				},
				CipherSuites: utils.CipherSuites,
				Certificates: []tls.Certificate{
					certificate,
				},
			},
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
		},
	}
}

// Start handles everything needed to start http-server daemon
func (server Server) Start() {
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
		defer server.Stop()
		log.Infof("Start http-server daemon, listening on %s", server.underlying.Addr)
		tlsListener := tls.NewListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, server.underlying.TLSConfig)
		err := server.underlying.Serve(tlsListener)
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("http-server error %v", err)
		}
	}()

	go func() {
		<-server.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		server.underlying.Shutdown(ctx)
		cancel()
		server.MarkDone()
		return
	}()

	server.WaitStop()
	log.Info("Stop http-server daemon")
}
