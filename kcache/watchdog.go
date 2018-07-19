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
	"fmt"
	"log"
	"time"

	"github.com/czcorpus/kontext-atn/taskdb"
)

const (
	watchdogWatchIntervalSec = 1
)

// ConcCacheEvent contains status data related
// to a specific concordance calculation.
type ConcCacheEvent struct {
	CorpusID string

	CacheKey string

	Record *taskdb.CacheRecord

	Error error
}

func (c *ConcCacheEvent) String() string {
	return fmt.Sprintf("ConcCacheEvent[CorpusID: %s, CacheKey: %s, Error: %s, Record: %v ]",
		c.CorpusID, c.CacheKey, c.Error, c.Record)
}

// ConcSize returns current conconrdance size
func (c *ConcCacheEvent) ConcSize() int {
	if c.Record != nil {
		return c.Record.ConcSize
	}
	return -1
}

// RelConcSize returns a relative concordance
// size scaled to million tokens (aka "i.p.m")
func (c *ConcCacheEvent) RelConcSize() float32 {
	if c.Record != nil {
		return c.Record.RelConcSize
	}
	return -1
}

// Finished returns calc. status ("true" means "finished")
func (c *ConcCacheEvent) Finished() bool {
	if c.Record != nil {
		return c.Record.Finished
	}
	return true
}

// FullSize - this value has no clear use in KonText
// but we keep it passing around.
func (c *ConcCacheEvent) FullSize() int {
	if c.Record != nil {
		return c.Record.FullSize
	}
	return -1
}

// Watch looks in regular intervals for a specific cache key and
// sends the data via 'events' channel.
func Watch(cacheDb *taskdb.ConcCacheDB, corpusID string, cacheKey string, events chan *ConcCacheEvent) {

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
		time.Sleep(time.Duration(watchdogWatchIntervalSec) * time.Second)

	}
	log.Printf("Watchdog for cache item %s finished.", cacheKey)
}
