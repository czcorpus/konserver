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
	"io"
	"log"
	"net/http"
	"path/filepath"

	"github.com/czcorpus/kontext-atn/kcache"
	"github.com/gorilla/websocket"
)

// Config defines a configuration
// required by kontext-atn to run the
// embedded WebSocket server.
type Config struct {
	Address string `json:"address"`
}

// WSServer handles HTTP/WebSocket requests/connections defined for kontex-atn
type WSServer struct {
	conf          *Config
	hub           *Hub
	cacheRootPath string
}

// NewWSServer creates a properly initialized
// instance of WSServer
func NewWSServer(hub *Hub, conf *Config, cacheRootPath string) *WSServer {
	ans := &WSServer{
		conf:          conf,
		hub:           hub,
		cacheRootPath: cacheRootPath,
	}
	ans.init()
	return ans
}

func (s *WSServer) init() {
	http.HandleFunc("/", s.serveHome)
	http.HandleFunc("/ws", s.serveNotifier)
}

// Serve starts the server and blocks until
// it is closed.
func (s *WSServer) Serve() {
	log.Printf("INFO: Listening on %s", s.conf.Address)
	http.ListenAndServe(s.conf.Address, nil)
}

func (s *WSServer) serveNotifier(writer http.ResponseWriter, request *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // TODO !!!
		},
	}
	conn, err := upgrader.Upgrade(writer, request, nil)

	corpusID := request.URL.Query().Get("corpusId")
	cacheKey := request.URL.Query().Get("cacheKey")
	cacheIdent := &kcache.CacheIdent{
		CorpusID:      corpusID,
		CacheKey:      cacheKey,
		CacheFilePath: filepath.Join(s.cacheRootPath, corpusID, cacheKey+".conc"),
	}
	s.hub.Register <- NewClient(cacheIdent, s.hub, conn)

	if err != nil {
		log.Println(err)
		return
	}
}

func (s *WSServer) serveHome(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
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
