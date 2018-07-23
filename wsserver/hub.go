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
	"crypto/md5"
	"fmt"
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
	watchedTasks    map[string]*Client // cache ID => client
	lastUpdates     map[string]int64   // cache ID => unix time
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
		lastUpdates:     make(map[string]int64),
		cacheDB:         cacheDB,
	}
}

func mkClientHash(client *Client) string {
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

// Run starts listen on all the channels.
// This must run in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.watchedTasks[mkClientHash(client)] = client
			log.Printf("INFO: Registered %v", client)
			go client.Run()
			go kcache.Watch(h.cacheDB, client.CacheIdent().CorpusID, client.CacheIdent().CacheKey, h.ConcCacheEvents)
		case client := <-h.Unregister:
			if _, ok := h.watchedTasks[mkClientHash(client)]; ok {
				delete(h.watchedTasks, mkClientHash(client))
			}
			log.Printf("INFO: Unregistered %v", client)
		case event := <-h.ConcCacheEvents:
			eventHash := mkEventHash(event)
			client, ok := h.watchedTasks[eventHash]
			if ok {
				if event.Error != nil {
					client.Errors <- event.Error

				} else if event.Record.LastUpdate > h.lastUpdates[eventHash] {
					status := NewConcStatusResponse(event)
					client.Incoming <- status
					h.lastUpdates[eventHash] = event.Record.LastUpdate
				}
			}
		}

	}
}
