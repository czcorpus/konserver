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
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/czcorpus/konserver/kcache"
	"github.com/gorilla/websocket"
)

// Client keeps connection with actual remote
// client (= browser) and sends data it recieves
// from a respective channel.
type Client struct {
	cacheIdent *kcache.CacheIdent
	hub        *Hub
	conn       *websocket.Conn
	Incoming   chan *kcache.ConcCacheEvent
	lastUpdate int64
	stop       chan bool
}

// NewClient creates a proper instance of Client
// with all the channels initialized.
func NewClient(cacheIdent *kcache.CacheIdent, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		cacheIdent: cacheIdent,
		hub:        hub,
		conn:       conn,
		Incoming:   make(chan *kcache.ConcCacheEvent),
		lastUpdate: 0,
		stop:       make(chan bool),
	}
}

func (c *Client) String() string {
	return fmt.Sprintf("Client[%s, %s]", c.cacheIdent.CorpusID, c.cacheIdent.CacheKey)
}

// CacheIdent returns a complete concordance cache
// identification based on how KonText works.
func (c *Client) CacheIdent() *kcache.CacheIdent {
	return c.cacheIdent
}

// Stop asynchronously stops the client
// by sending 'true' to a respective channel.
func (c *Client) Stop() {
	c.stop <- true
}

// Run starts to listen on all the channels.
// This method must be used within a goroutine.
func (c *Client) Run() {
	defer c.conn.Close()
	for {
		cw, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Fatal("ERROR: Failed to create message writer ", err)
		}
		select {
		case stop := <-c.stop:
			if stop {
				return
			}
		case event := <-c.Incoming:
			if event.Error != nil {
				c.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseInternalServerErr,
						fmt.Sprintf("%s", event.Error)))

			} else if event.Record.LastUpdate > c.lastUpdate {
				status := NewConcStatusResponse(event)
				c.lastUpdate = event.Record.LastUpdate
				ans, err := json.Marshal(status)
				if err != nil {
					c.conn.WriteMessage(websocket.CloseMessage,
						websocket.FormatCloseMessage(websocket.CloseInternalServerErr,
							fmt.Sprintf("%s", err)))
				}
				cw.Write(ans)
				if event.Record.Finished {
					c.conn.WriteMessage(websocket.CloseMessage,
						websocket.FormatCloseMessage(websocket.CloseNormalClosure, "DONE"))
					c.hub.Unregister <- c
					return
				}
			}
		case <-time.After(1 * time.Minute):
			log.Printf("INFO: Closing client for cache item %s after timeout.", c.cacheIdent.CacheKey)
			return
		}
	}
}
