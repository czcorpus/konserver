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

package workpool

import (
	"fmt"
	"time"
)

const (
	// taskStatusWaiting means not executed yet
	taskStatusWaiting = 0

	// taskStatusRunning means execution started
	// and no end status obtained yet
	taskStatusRunning = 1

	// taskStatusFinished means execution finished
	// without error
	taskStatusFinished = 2
)

type Task struct {
	TaskID  string      `json:"taskID"`
	Status  int         `json:"status"`
	Fn      string      `json:"fn"`
	Args    interface{} `json:"args"`
	Error   string      `json:"error"`
	Result  interface{} `json:"result"`
	Created int64       `json:"created"`
	Updated int64       `json:"updated"`
}

func (t *Task) IsDone() bool {
	return t.Status == taskStatusFinished
}

func (t *Task) String() string {
	return fmt.Sprintf("Task[id: %s, created: %d, status: %d, fn: %s, error: %s, args: %v",
		t.TaskID, t.Created, t.Status, t.Fn, t.Error, t.Args)
}

func (t *Task) AgeSecons() int {
	return int(time.Now().Unix() - t.Created)
}

func (t *Task) SecondsSinceUpdate() int {
	return int(time.Now().Unix() - t.Updated)
}

func (t *Task) Touch() {
	t.Updated = time.Now().Unix()
}
