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
)

// Hub controls the communication between
// calculation watchdogs and WebSocket clients.
type Hub struct {
	Register        chan *Client
	Unregister      chan *Client
	ConcCacheEvents chan *kcache.ConcCacheEvent
	watchedTasks    map[string]*Client // task id => client
}

// NewHub creates a proper instance of the Hub
// with all the channels initialized
func NewHub() *Hub {
	return &Hub{
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		ConcCacheEvents: make(chan *kcache.ConcCacheEvent, 5), // TODO configurable buffer
		watchedTasks:    make(map[string]*Client),
	}
}

// Run starts listen on all the channels.
// This must run in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.watchedTasks[client.taskID] = client
			log.Printf("INFO: Hub registered task %s", client.taskID)
			go client.Run()
			go kcache.Watch(h.ConcCacheEvents, client.taskID)
		case client := <-h.Unregister:
			if _, ok := h.watchedTasks[client.taskID]; ok {
				delete(h.watchedTasks, client.taskID)
			}
			log.Printf("INFO: Hub unregistered task %s", client.taskID)
		case event := <-h.ConcCacheEvents:
			client, ok := h.watchedTasks[event.TaskID]
			status := NewConcStatusResponse(event)
			if ok {
				client.Incoming <- status
			}
		}

	}
}
