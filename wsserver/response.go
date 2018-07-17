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
	"github.com/czcorpus/kontext-atn/kcache"
)

type ConcStatusResponse struct {
	FullSize            int      `json:"fullsize"`
	ConcSize            int      `json:"concsize"`
	RelConcSize         float32  `json:"relconcsize"`
	ConcPersistenceOpID string   `json:"conc_persistence_op_id"`
	Messages            []string `json:"messages"`
	UserOwnsConc        bool     `json:"user_owns_conc"`
	Q                   []string `json:"Q"`
	Finished            bool     `json:"finished"`
	ARF                 float32  `json:"arf"`
}

func NewConcStatusResponse(evt *kcache.ConcCacheEvent) *ConcStatusResponse {
	return &ConcStatusResponse{
		ConcSize:    evt.ConcSize,
		RelConcSize: evt.RelConcSize,
		Finished:    evt.Finished,
	}
}
