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

package nullqueue

import (
	"log"

	"github.com/czcorpus/konserver/workpool"
)

// NullQueue is a dummy replacement for
// an actual task queue. It is used in
// case konserver is run without configured
// queue (in such case it works as a helper
// server for Celery to notify users about
// conc. calculation status via WebSocket)
type NullQueue struct {
}

// Info returns overview information used
// on the "info" page of the API server.
func (nq *NullQueue) Info() *workpool.MasterInfo {
	return &workpool.MasterInfo{
		PoolSize: 0,
	}
}

// GetTask returns always nil
func (nq *NullQueue) GetTask(taskID string) *workpool.Task {
	return nil
}

// SendTask fakes creating a new task.
// The function has no effect.
func (nq *NullQueue) SendTask(name string, jsonArgs []byte) *workpool.Task {
	return nil
}

// Start fakes starting the service.
// The function has no effect.
func (nq *NullQueue) Start() {
	log.Print("WARNING: Worker server is disabled in the configuration")
}

// Stop fakes stopping the service.
func (nq *NullQueue) Stop() {}
