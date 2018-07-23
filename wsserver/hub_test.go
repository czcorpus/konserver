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
	"testing"

	"github.com/czcorpus/kontext-atn/kcache"
	"github.com/stretchr/testify/assert"
)

func TestMkClientHash(t *testing.T) {
	c := &Client{
		cacheIdent: &kcache.CacheIdent{
			CorpusID: "foo",
			CacheKey: "abcdef",
		},
	}
	h := mkClientHash(c)
	assert.Equal(t, "332b91ee74f70e2999c68cb513102a2b", h)
}

func TestMKEventHash(t *testing.T) {
	e := &kcache.ConcCacheEvent{
		CorpusID: "foo",
		CacheKey: "abcdef",
	}
	h := mkEventHash(e)
	assert.Equal(t, "332b91ee74f70e2999c68cb513102a2b", h)
}

func TestMkClientEventHashEqual(t *testing.T) {
	c := &Client{
		cacheIdent: &kcache.CacheIdent{
			CorpusID: "syn2015",
			CacheKey: "abcdef01234567890",
		},
	}
	h1 := mkClientHash(c)
	e := &kcache.ConcCacheEvent{
		CorpusID: "syn2015",
		CacheKey: "abcdef01234567890",
	}
	h2 := mkEventHash(e)
	assert.True(t, h1 == h2)
}
