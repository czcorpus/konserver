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

	"github.com/gorilla/websocket"
)

// Client keeps connection with actual remote
// client (= browser) and sends data it recieves
// from a respective channel.
type Client struct {
	CorpusID string
	CacheKey string
	hub      *Hub
	conn     *websocket.Conn
	Incoming chan *ConcStatusResponse
}

// NewClient creates a proper instance of Client
// with all the channels initialized.
func NewClient(corpusID string, cacheKey string, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		CorpusID: corpusID,
		CacheKey: cacheKey,
		hub:      hub,
		conn:     conn,
		Incoming: make(chan *ConcStatusResponse),
	}
}

// Run starts to listen on all the channels.
// This method must be used within a goroutine.
func (c *Client) Run() {
	defer func() {
		c.hub.Unregister <- c
		c.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(1000, "DONE"))
		c.conn.Close()
	}()

	for {
		cw, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			log.Fatal("ERROR: Failed to create message writer ", err)
		}
		select {
		case msg := <-c.Incoming:
			ans, err := json.Marshal(msg)
			if err != nil {
				c.conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(1011, fmt.Sprintf("%s", err)))
			}
			cw.Write(ans)
		case <-time.After(5 * time.Second):
			log.Printf("INFO: Closing client for cache item %s after timeout.", c.CacheKey)
			return
		}
	}
}
