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
)

// ConcCacheEvent contains status data related
// to a specific concordance calculation.
type ConcCacheEvent struct {

	// A respective Celery (or other worker queue) task identifier
	TaskID string

	// The current concordance size (as the calculation runs)
	ConcSize int

	// ??? TODO
	RelConcSize float32

	// Calculation status (true should be set in any finished state - incl. error)
	Finished bool
}

// Watch is currently a fake stuff to be able
// to test WebSocket & Hub part
func Watch(events chan *ConcCacheEvent, taskID string) {
	i := 0
	for {
		finished := false
		if i == 4 {
			finished = true
		}
		events <- &ConcCacheEvent{
			TaskID:      taskID,
			ConcSize:    i,
			RelConcSize: 0.0,
			Finished:    finished,
		}

		if finished {
			break
		}
		time.Sleep(time.Duration(1) * time.Second)
		i++
	}
	log.Printf("Watchdog for task %s finished.", taskID)
}
