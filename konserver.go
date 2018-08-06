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
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/czcorpus/konserver/apiserver"
	"github.com/czcorpus/konserver/taskdb"
	"github.com/czcorpus/konserver/workpool"
	"github.com/czcorpus/konserver/workpool/nullqueue"
)

func main() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP)
	flag.Parse()

	for {
		conf, err := loadConfig(flag.Arg(0))
		if err != nil {
			log.Fatalf("ERROR: Failed to read conf %s: %s", flag.Arg(0), err)
		}
		if conf.LogPath != "" {
			logf, err := os.OpenFile(conf.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
			if err != nil {
				log.Fatal("ERROR: ", err)
			}
			log.SetOutput(logf)
		}

		cacheDB := taskdb.NewConcCacheDB(&conf.Redis)
		hub := apiserver.NewHub(cacheDB)
		var taskMaster apiserver.TaskMaster
		if conf.WorkerMaster.PoolSize > 0 {
			taskMaster = workpool.NewMaster(&conf.WorkerMaster)

		} else {
			taskMaster = &nullqueue.NullQueue{}
		}
		server := apiserver.NewAPIServer(hub, &conf.APIServerConfig, taskMaster, conf.CacheRootDir)

		go hub.Run()
		go server.Serve()
		go taskMaster.Start()

		<-sc
		log.Print("Reloading services...")
		server.Shutdown()
	}
}
