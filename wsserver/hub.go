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

	"github.com/czcorpus/kontext-atn/kcache"
	"github.com/czcorpus/kontext-atn/taskdb"
)

// Hub controls the communication between
// calculation watchdogs and WebSocket clients.
type Hub struct {
	Register        chan *Client
	Unregister      chan *Client
	ConcCacheEvents chan *kcache.ConcCacheEvent
	watchedTasks    map[string]*Client // task id => client
	cacheDB         *taskdb.ConcCacheDB
}

// NewHub creates a proper instance of the Hub
// with all the channels initialized
func NewHub(cacheDB *taskdb.ConcCacheDB) *Hub {
	return &Hub{
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		ConcCacheEvents: make(chan *kcache.ConcCacheEvent, 5), // TODO configurable buffer
		watchedTasks:    make(map[string]*Client),
		cacheDB:         cacheDB,
	}
}

// Run starts listen on all the channels.
// This must run in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.watchedTasks[client.CacheKey] = client
			log.Printf("INFO: Hub registered task %s", client.CacheKey)
			go client.Run()
			go kcache.Watch(h.ConcCacheEvents, h.cacheDB, client.CorpusID, client.CacheKey)
		case client := <-h.Unregister:
			if _, ok := h.watchedTasks[client.CacheKey]; ok {
				delete(h.watchedTasks, client.CacheKey)
			}
			log.Printf("INFO: Hub unregistered cache key %s", client.CacheKey)
		case event := <-h.ConcCacheEvents:
			client, ok := h.watchedTasks[event.CacheKey]
			status := NewConcStatusResponse(event)
			if ok {
				client.Incoming <- status
			}
		}

	}
}
