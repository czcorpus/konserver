package taskdb

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
)

type ConcCacheDBConf struct {
	Address  string `json:"address"`
	Database int    `json:"database"`
}

type ConcCacheDB struct {
	db *redis.Client
}

func NewConcCacheDB(conf *ConcCacheDBConf) *ConcCacheDB {
	return &ConcCacheDB{
		db: redis.NewClient(&redis.Options{
			Addr:     conf.Address,
			Password: "",
			DB:       conf.Database,
		}),
	}
}

func (c *ConcCacheDB) GetItem(corpusId string, cacheKey string) (*CacheRecord, error) {
	rawRec := c.db.HGet(fmt.Sprintf("conc_cache:%s", corpusId), cacheKey)
	log.Printf("RAW REC[%s %s]: %s", corpusId, cacheKey, rawRec)
	return parse(rawRec.Val())
}
