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

package apiserver

import (
	"crypto/md5"
	"fmt"
	"log"

	"github.com/czcorpus/konserver/kcache"
	"github.com/czcorpus/konserver/taskdb"
)

// Watcher represents an autonomous object
// watching for changes in concordance calculation.
// Here we're interested only in how to start
// and stop the object.
type Watcher interface {
	Start()
	Stop()
}

func mkClientHash(client *WSClient) string {
	h := md5.New()
	h.Write([]byte(client.CacheIdent().CorpusID))
	h.Write([]byte(client.CacheIdent().CacheKey))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func mkEventHash(evt *kcache.ConcCacheEvent) string {
	h := md5.New()
	h.Write([]byte(evt.CorpusID))
	h.Write([]byte(evt.CacheKey))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Hub controls the communication between
// calculation watchdogs and WebSocket clients.
type Hub struct {
	Register        chan *WSClient
	Unregister      chan *WSClient
	stop            chan bool
	watchdogFactory *kcache.RedisWatchdogFactory
	clients         map[string]*WSClient // cache ID => client
	watchdogs       map[string]Watcher
	cacheDB         *taskdb.ConcCacheDB
}

// NewHub creates a proper instance of the Hub
// with all the channels initialized
func NewHub(cacheDB *taskdb.ConcCacheDB) *Hub {
	return &Hub{
		watchdogFactory: kcache.NewRedisWatchdogFactory(cacheDB),
		Register:        make(chan *WSClient),
		Unregister:      make(chan *WSClient),
		stop:            make(chan bool, 1),
		watchdogs:       make(map[string]Watcher),
		clients:         make(map[string]*WSClient),
		cacheDB:         cacheDB,
	}
}

// Run starts listen on all the channels.
// This must run in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case <-h.stop: // stop whole hub along with all the registered clients & watchdogs
			for _, w := range h.watchdogs {
				w.Stop()
			}
			for _, c := range h.clients {
				c.Stop()
			}
			return
		case client := <-h.Register:
			key := mkClientHash(client)
			h.clients[key] = client
			log.Printf("INFO: Registered %v", client)
			go client.Run()
			h.watchdogs[key] = h.watchdogFactory.Create(client.CacheIdent(), client.Incoming)
			go h.watchdogs[key].Start()
		case client := <-h.Unregister:
			key := mkClientHash(client)
			if w, ok := h.watchdogs[key]; ok {
				w.Stop()
				delete(h.watchdogs, key)
			}
			if _, ok := h.clients[key]; ok {
				delete(h.clients, key)
			}
			log.Printf("INFO: Unregistered %v", client)
		}
	}
}
