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
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/satori/go.uuid"
)

// MasterConf is a Master configuration
type MasterConf struct {

	// PoolSize specifies number of workers
	PoolSize int `json:"poolSize"`

	// Program specifies program name (basically the first
	// element of a command we want to use as a worker)
	Program string `json:"program"`

	// PogramArgs specifies all other command arguments
	// ("the rest" in list terminology)
	ProgramArgs []string `json:"programArgs"`

	// ExecMaxSeconds specifies a time a task have
	// to execute (i.e. we start to count once the
	// task is actually started - not equeued).
	ExecMaxSeconds int `json:"execMaxSeconds"`

	TaskResultPersistMaxSeconds int `json:"taskResultPersistMaxSeconds"`
}

type MasterInfo struct {
	PoolSize    int
	WorkersInfo []WorkerInfo
}

// Master handles distribution of tasks to workers
// and also collecting results.
type Master struct {
	conf        *MasterConf
	workers     map[*Worker]*Task
	tasks       map[string]*Task
	queue       chan *Task
	queueEvent  chan bool // true = new item, false = removed item
	workerEvent chan *WorkerStatus
	mutex       *sync.Mutex
}

// NewMaster is a standard constructor for Master
func NewMaster(conf *MasterConf) *Master {
	return &Master{
		conf:        conf,
		workers:     make(map[*Worker]*Task),
		tasks:       make(map[string]*Task),
		queueEvent:  make(chan bool, conf.PoolSize*10),
		queue:       make(chan *Task, conf.PoolSize),
		workerEvent: make(chan *WorkerStatus, conf.PoolSize*10),
		mutex:       &sync.Mutex{},
	}
}

// Info returns overview information used
// on the "info" page of the API server.
func (m *Master) Info() *MasterInfo {
	workersInfo := make([]WorkerInfo, len(m.workers))
	i := 0
	for worker := range m.workers {
		workersInfo[i] = worker.Info()
		i++
	}
	return &MasterInfo{
		PoolSize:    len(m.workers),
		WorkersInfo: workersInfo,
	}
}

// getFreeWorker returns a free Worker if available.
// Otherwise, nil is returned.
func (m *Master) getFreeWorker() *Worker {
	for w, t := range m.workers {
		if t == nil {
			return w
		}
	}
	return nil
}

// executeNextTask fetches a next task from
// internal queue and executes it. In case
// there is no task enqueued, nothing is done.
func (m *Master) executeNextTask() {
	worker := m.getFreeWorker()
	if worker != nil {
		select {
		case task := <-m.queue:
			log.Print("INFO: dequed task ", task)
			if task != nil {
				m.workers[worker] = task
				task.Status = taskStatusRunning
				worker.Call(task.TaskID, task.Fn, task.Args)
			}
		default:
		}
	}
}

func (m *Master) checkForStuckWorkers() {
	for worker, task := range m.workers {
		if task != nil && task.AgeSecons() > m.conf.ExecMaxSeconds {
			log.Print("checking task ", time.Now().Unix(), task.Created, task.AgeSecons(), m.conf.ExecMaxSeconds)
			//m.mutex.Lock()
			m.workers[worker] = nil
			worker.Stop() // TODO what if this takes a long time???
			worker.Start()
			task.Error = "Task execution limit reached"
			task.Status = taskStatusFinished
			task.Touch()
			//m.mutex.Unlock()
			log.Printf("WARNING: restarted stuck worker %v", worker)
			m.queueEvent <- true
		}
	}
}

func (m *Master) checkForOldTasks() {
	for taskID, task := range m.tasks {
		if task.IsDone() && task.SecondsSinceUpdate() > m.conf.TaskResultPersistMaxSeconds {
			delete(m.tasks, taskID)
		}
	}
}

// listenForEvents starts a goroutine listening
// for changes Master interprets as triggers
// to start another task (e.g. "new task has been
// added", "existing task has finished").
func (m *Master) listenForEvents() {
	go func() {
		for {
			select {
			case v := <-m.queueEvent:
				if v {
					m.executeNextTask()

				} else {
					log.Print("INFO: removed task")
				}
			case v := <-m.workerEvent:
				if v.IsDone() {
					task, ok := m.tasks[v.TaskID]
					if !ok {
						// TODO
						log.Print("ERROR: worker event no longer valid (task gone)")
					}
					if v.IsDone() {
						task.Error = v.Error
						task.Status = taskStatusFinished
						task.Result = v.Result
						m.workers[v.Worker()] = nil
					}
					task.Touch()
					m.queueEvent <- true

				} else {
					log.Print("INFO: updated status of worker", v)
					// TODO
				}
			case <-time.After(1 * time.Second):
				m.checkForStuckWorkers()
				m.checkForOldTasks()
			}
		}
	}()
}

// Start initializes all the worker processes
// and starts to listen for tasks. The function
// is non-blocking.
func (m *Master) Start() {
	for i := 0; i < m.conf.PoolSize; i++ {
		args := append(m.conf.ProgramArgs, fmt.Sprintf("W%d", i))
		worker := NewWorker(m.workerEvent, m.conf.Program, args...)
		m.workers[worker] = nil
		worker.Start() // TODO we must catch errors in worker via a channel
		/*
			if werr != nil {
				m.Stop()
				log.Printf("ERROR: failed to run worker %d: %s", i, werr) // TODO
			}
		*/
		log.Printf("INFO: started worker %v", worker)
	}
	m.listenForEvents()
}

// Stop stops all the workers
func (m *Master) Stop() {
	for w := range m.workers {
		if w != nil {
			w.Stop()
		}
	}
	// TODO stop also listener for tasks etc.
}

// Reload reloads Master and all the workers.
// This can be used to update service configuration.
func (m *Master) Reload() {
	for w := range m.workers {
		w.Reload()
	}
}

// GetTask returns a specific task identified
// by task ID. In case there is no such task,
// nil is returned.
func (m *Master) GetTask(taskID string) *Task {
	for _, task := range m.tasks {
		if task.TaskID == taskID {
			return task
		}
	}
	return nil
}

// SendTask sends a new task to Master
func (m *Master) SendTask(name string, jsonArgs []byte) *Task {
	log.Printf("Received task %s with args %s", name, string(jsonArgs))
	taskID, err := uuid.NewV4()
	if err != nil {
		// TODO
		log.Print("ERROR: ", err)
	}
	var args interface{}
	err = json.Unmarshal(jsonArgs, &args)
	if err != nil {
		log.Print("ERROR: ", err)
	}
	task := &Task{
		TaskID:  taskID.String(),
		Status:  taskStatusWaiting,
		Fn:      name,
		Args:    args,
		Created: time.Now().Unix(),
	}
	m.tasks[task.TaskID] = task
	m.queue <- task
	log.Print("INFO: >>>> ENQUEUED TASK ", task)
	m.queueEvent <- true
	return task
}
