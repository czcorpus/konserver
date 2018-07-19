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

package kcache

import (
	"log"
	"time"

	"github.com/czcorpus/kontext-atn/taskdb"
)

// ConcCacheEvent contains status data related
// to a specific concordance calculation.
type ConcCacheEvent struct {
	CorpusID string

	CacheKey string

	Record *taskdb.CacheRecord

	Error error
}

func (c *ConcCacheEvent) ConcSize() int {
	return c.Record.ConcSize
}

func (c *ConcCacheEvent) RelConcSize() float32 {
	return c.Record.RelConcSize
}

func (c *ConcCacheEvent) Finished() bool {
	return c.Record.Finished
}

func (c *ConcCacheEvent) FullSize() int {
	return c.Record.FullSize
}

// Watch is currently a fake stuff to be able
// to test WebSocket & Hub part
func Watch(events chan *ConcCacheEvent, cacheDb *taskdb.ConcCacheDB, corpusID string, cacheKey string) {

	for {

		rec, err := cacheDb.GetItem(corpusID, cacheKey)
		if err != nil {
			events <- &ConcCacheEvent{
				CorpusID: corpusID,
				CacheKey: cacheKey,
				Record:   rec,
				Error:    err,
			}
			break
		}

		events <- &ConcCacheEvent{
			CorpusID: corpusID,
			CacheKey: cacheKey,
			Record:   rec,
		}
		if rec.Finished {
			break
		}
		time.Sleep(time.Duration(1) * time.Second)

	}
	log.Printf("Watchdog for cache item %s finished.", cacheKey)
}
