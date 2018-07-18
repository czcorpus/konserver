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
	"flag"
	"io/ioutil"
	"log"

	"github.com/czcorpus/kontext-atn/taskdb"
	"github.com/czcorpus/kontext-atn/wsserver"
)

type AppConfig struct {
	WSServerConfig wsserver.WSServerConfig `json:"wsServer"`
	Redis          taskdb.ConcCacheDBConf  `json:"cacheDb"`
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

func main() {
	flag.Parse()
	conf, err := loadConfig(flag.Arg(0))
	if err != nil {
		log.Fatalf("Failed to read conf %s: %s", flag.Arg(0), err)
	}
	cacheDB := taskdb.NewConcCacheDB(&conf.Redis)
	hub := wsserver.NewHub(cacheDB)
	go hub.Run()
	server := wsserver.NewWSServer(hub, &conf.WSServerConfig)
	server.Serve()
}
