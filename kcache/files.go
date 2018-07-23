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
	"os"

	"github.com/fsnotify/fsnotify"
)

func errEvent(cacheIdent *CacheIdent, err error) *ConcCacheEvent {
	return &ConcCacheEvent{
		CorpusID: cacheIdent.CorpusID,
		CacheKey: cacheIdent.CacheKey,
		Error:    err,
	}
}

func WatchFile(cacheIdent *CacheIdent, events chan *ConcCacheEvent) {
	_, err := os.Stat(cacheIdent.CacheFilePath)
	log.Print("##### WATCH CACHE FILE ", cacheIdent.CacheFilePath, err)
	if err != nil {
		events <- errEvent(cacheIdent, err)
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		events <- errEvent(cacheIdent, err)
	}
	defer watcher.Close()

	go func() {
		log.Printf("INFO: watching file %s", cacheIdent.CacheFilePath)
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err := <-watcher.Errors:
				events <- errEvent(cacheIdent, err)
			}
		}
	}()

	err = watcher.Add(cacheIdent.CacheFilePath)
	if err != nil {
		events <- errEvent(cacheIdent, err)
	}
}
