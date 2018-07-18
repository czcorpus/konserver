package taskdb

import (
	"encoding/json"
)

type CacheRecord struct {
	CurrWait    float32 `json:"curr_wait"`
	TaskID      string  `json:"task_id"`
	Created     int     `json:"created"`
	LastUpdate  int     `json:"last_upd"`
	PID         int     `json:"pid"`
	RelConcSize float32 `json:"relconcsize"`
	FullSize    int     `json:"fullsize"`
	Finished    bool    `json:"finished"`
	Error       string  `json:"error"`
	ConcSize    int     `json:"concsize"`
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
