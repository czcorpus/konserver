// Copyright 2018 Tomas Machalek <tomas.machalek@gmail.com>
// Copyright (c) 2018 Charles University, Faculty of Arts,
//                    Institute of the Czech National Corpus
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wsserver

import (
	"context"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/czcorpus/kontext-atn/kcache"
	"github.com/gorilla/websocket"
)

// Config defines a configuration
// required by kontext-atn to run the
// embedded WebSocket server.
type Config struct {
	Address        string   `json:"address"`
	URLPathRoot    string   `json:"urlPathRoot"`
	AllowedOrigins []string `json:"allowedOrigins"`
	SSLCertFile    string   `json:"sslCertFile"`
	SSLKeyFile     string   `json:"sslKeyFile"`
}

// WSServer handles HTTP/WebSocket requests/connections defined for kontex-atn
type WSServer struct {
	conf          *Config
	httpServer    *http.Server
	mux           *http.ServeMux
	hub           *Hub
	cacheRootPath string
}

// NewWSServer creates a properly initialized
// instance of WSServer
func NewWSServer(hub *Hub, conf *Config, cacheRootPath string) *WSServer {
	mux := http.NewServeMux()
	ans := &WSServer{
		conf:          conf,
		mux:           mux,
		httpServer:    &http.Server{Addr: conf.Address, Handler: mux},
		hub:           hub,
		cacheRootPath: cacheRootPath,
	}

	if !strings.HasPrefix(ans.conf.URLPathRoot, "/") {
		log.Fatal("URLPathRoot must start with /")
	}
	ans.mux.HandleFunc(conf.URLPathRoot+"/", ans.serveHome)
	ans.mux.HandleFunc(conf.URLPathRoot+"/ws", ans.serveNotifier)

	return ans
}

// Serve starts the server and blocks until
// it is closed.
func (s *WSServer) Serve() {
	log.Printf("INFO: Serving at %s", s.conf.Address+s.conf.URLPathRoot)
	if s.conf.SSLCertFile != "" && s.conf.SSLKeyFile != "" {
		s.httpServer.ListenAndServeTLS(s.conf.SSLCertFile, s.conf.SSLKeyFile)
	} else {
		s.httpServer.ListenAndServe()
	}
	http.ListenAndServe(s.conf.Address, s.mux)
}

// Shutdown gracefully stops the server
func (s *WSServer) Shutdown() {
	log.Print("INFO: Shutting down the web server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.httpServer.Shutdown(ctx)
}

func (s *WSServer) serveNotifier(writer http.ResponseWriter, request *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			for _, allowedOrigin := range s.conf.AllowedOrigins {
				if origin == allowedOrigin {
					return true
				}
			}
			log.Printf("ERROR: origin %s not found in allowed origins list", origin)
			return false
		},
	}
	conn, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		log.Print("ERROR: ", err)
		return
	}

	corpusID := request.URL.Query().Get("corpusId")
	cacheKey := request.URL.Query().Get("cacheKey")
	cacheIdent := &kcache.CacheIdent{
		CorpusID:      corpusID,
		CacheKey:      cacheKey,
		CacheFilePath: filepath.Join(s.cacheRootPath, corpusID, cacheKey+".conc"),
	}
	s.hub.Register <- NewClient(cacheIdent, s.hub, conn)
}

// serveHome provides some information about running server
func (s *WSServer) serveHome(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != s.conf.URLPathRoot+"/info" {
		http.Error(writer, "Not found", http.StatusNotFound)
		return
	}
	if request.Method != "GET" {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	out := "This is kontext-atn WebSocket server.\n\nUse /ws?corpusId=...&cacheKey=...\nto use concordance status notification service."
	writer.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	io.WriteString(writer, out)
}
