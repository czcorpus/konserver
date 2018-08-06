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

package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/czcorpus/konserver/apiserver"
	"github.com/czcorpus/konserver/taskdb"
	"github.com/czcorpus/konserver/workpool"
)

// AppConfig contains whole konserver
// configuration
type AppConfig struct {
	APIServerConfig apiserver.Config       `json:"apiServer"`
	Redis           taskdb.ConcCacheDBConf `json:"cacheDb"`
	CacheRootDir    string                 `json:"cacheRootDir"`
	WorkerMaster    workpool.MasterConf    `json:"workerMaster"`
	LogPath         string                 `json:"logPath"`
}

// ConfiguresQueue tests whether the application
// is configured to run in "task queue" mode
// (i.e. as a Celery replacement).
func (ac *AppConfig) ConfiguresQueue() bool {
	return ac.WorkerMaster.PoolSize > 0
}

func loadConfig(path string) (*AppConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var conf AppConfig
	err = json.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
