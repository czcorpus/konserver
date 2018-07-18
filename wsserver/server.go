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
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type WSServerConfig struct {
	Address string `json:"address"`
}

type WSServer struct {
	conf   *WSServerConfig
	router *mux.Router
	hub    *Hub
}

func NewWSServer(hub *Hub, conf *WSServerConfig) *WSServer {
	ans := &WSServer{
		conf:   conf,
		hub:    hub,
		router: mux.NewRouter(),
	}
	ans.init()
	return ans
}

func (s *WSServer) init() {
	s.router.HandleFunc("/", s.serveHome)
	s.router.HandleFunc("/ws", s.serveNotifier)
	http.Handle("/", s.router)
}

func (s *WSServer) Serve() {
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
	s.hub.Register <- NewClient(corpusID, cacheKey, s.hub, conn)

	if err != nil {
		log.Println(err)
		return
	}
}

func (s *WSServer) serveHome(writer http.ResponseWriter, request *http.Request) {
	log.Print("Serve home: ", request.URL.Path)
	if request.URL.Path != "/" {
		http.Error(writer, "Not found", http.StatusNotFound)
		return
	}
	log.Print("cool")
	if request.Method != "GET" {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(writer, request, "./resources/index.html")
}
