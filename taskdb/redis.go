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
	"fmt"

	"github.com/go-redis/redis"
)

const (
	cacheKeyTemplate = "conc_cache:%s"
)

// ConcCacheDBConf specifies a configuration
// of a connection to a Redis database containing
// KonText cache records.
type ConcCacheDBConf struct {
	Address  string `json:"address"`
	Database int    `json:"database"`
}

// ConcCacheDB is a simple wrapper around Redis
// connection with interface customized/stripped to our
// needs.
type ConcCacheDB struct {
	db *redis.Client
}

// NewConcCacheDB creates a properly configured
// instance of ConcCacheDB
func NewConcCacheDB(conf *ConcCacheDBConf) *ConcCacheDB {
	return &ConcCacheDB{
		db: redis.NewClient(&redis.Options{
			Addr:     conf.Address,
			Password: "",
			DB:       conf.Database,
		}),
	}
}

// GetItem returns a cache item of a specific
// corpus and key.
func (c *ConcCacheDB) GetItem(corpusID string, cacheKey string) (*CacheRecord, error) {
	rawRec := c.db.HGet(fmt.Sprintf(cacheKeyTemplate, corpusID), cacheKey)
	return parse(rawRec.Val())
}
