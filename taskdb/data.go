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

package taskdb

import (
	"encoding/json"
	"fmt"
	"time"
)

// CacheRecord describes KonText's Redis cache record.
type CacheRecord struct {
	CurrWait    float32 `json:"curr_wait"`
	TaskID      string  `json:"task_id"`
	Created     int     `json:"created"`
	LastUpdate  int64   `json:"last_upd"`
	PID         int     `json:"pid"`
	RelConcSize float32 `json:"relconcsize"`
	ARF         float32 `json:"arf"`
	FullSize    int     `json:"fullsize"`
	Finished    bool    `json:"finished"`
	Error       string  `json:"error"`
	ConcSize    int     `json:"concsize"`
}

func (c *CacheRecord) String() string {
	return fmt.Sprintf("CacheRecord[PID: %d, TaskID: %s, Created: %s]", c.PID, c.TaskID, time.Unix(int64(c.Created), 0))
}

func parse(src string) (*CacheRecord, error) {
	tmp := make([]interface{}, 3)
	err := json.Unmarshal([]byte(src), &tmp)
	if err != nil {
		return nil, err
	}
	dataSrc, err := json.Marshal(tmp[1])
	if err != nil {
		return nil, err
	}
	var ans CacheRecord
	err = json.Unmarshal(dataSrc, &ans)
	if err != nil {
		return nil, err
	}
	return &ans, nil
}
