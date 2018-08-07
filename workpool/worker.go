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
	"os/exec"
	"syscall"
)

const (
	workerStatusIdle    = iota
	workerStatusRunning = iota
	workerStatusStopped = iota
)

// WorkerStatus describes current
// state and task (if applicable) info.
type WorkerStatus struct {
	TaskID string      `json:"taskID"`
	Status int         `json:"status"`
	Error  string      `json:"error"`
	Result interface{} `json:"result"`
	worker *Worker
}

func (ws *WorkerStatus) IsDone() bool {
	return ws.Status == workerStatusIdle || ws.Status == workerStatusStopped
}

func (ws *WorkerStatus) Worker() *Worker {
	return ws.worker
}

func (ws *WorkerStatus) ReadableStatus() string {
	switch ws.Status {
	case workerStatusRunning:
		return "running"
	case workerStatusIdle:
		return "idle"
	case workerStatusStopped:
		return "stopped"
	default:
		return "!unknown!"
	}
}

// ----------------------------------------------

type WorkerInfo struct {
	PID        int
	LastStatus string
	TaskID     string
}

// ----------------------------------------------

// Worker controls execution of an external task (program)
// via Cmd. The task must be able to receive commands via
// its standard input and response its status via its
// standard output. In general we suppose that each line
// in the streams represents either a command or a response.
// It means that the data must be encoded so possible new
// line characters which are part of commands do not split
// a single command into multiple commands.
type Worker struct {
	commandName               string
	args                      []string
	cmd                       *exec.Cmd
	commandsPipe              *CommandPipe
	responsesPipe             *ResponsePipe
	workerEvent               chan *WorkerStatus
	lastEvent                 WorkerStatus // this is used only for overview purposes
	taskID                    string
	maxResponsePipeBufferSize int
}

// workerCall describe a single function call
// to the worker. It is expected to be JSON
// serializable.
type workerCall struct {
	Fn     string      `json:"fn"`
	Args   interface{} `json:"args"`
	TaskID string      `json:"task_id"`
}

// NewWorker is a default factory for Worker
func NewWorker(workerEvent chan *WorkerStatus, maxResponsePipeBufferSize int, command string, args ...string) *Worker {
	return &Worker{
		commandName:               command,
		args:                      args,
		workerEvent:               workerEvent,
		maxResponsePipeBufferSize: maxResponsePipeBufferSize,
	}
}

func (w *Worker) String() string {
	return fmt.Sprintf("Worker[pid: %d, curr. task: %s]", w.GetPID(), w.taskID)
}

// GetPID returns actual PID of a respective external task.
// If nothing is running yet then -1 is returned.
func (w *Worker) GetPID() int {
	if w.cmd != nil && w.cmd.Process != nil {
		return w.cmd.Process.Pid
	}
	return -1
}

// Start runs the Worker - both communication in-memory pipes are
// set and the Worker is listening via a specific channel to
// responses of the task.
func (w *Worker) Start() {
	w.commandsPipe = NewCommandPipe()
	w.responsesPipe = NewResponsePipe(w.maxResponsePipeBufferSize)
	var err error
	w.cmd = exec.Command(w.commandName, w.args...)
	w.responsesPipe.Register(w.cmd)
	w.commandsPipe.Register(w.cmd)

	ch := w.responsesPipe.Channel()

	go func() {
		for {
			select {
			case data := <-ch:
				log.Print("GOT FROM PIPE ", data)
				var ans WorkerStatus
				var err error
				err = json.Unmarshal([]byte(data), &ans)
				log.Print("DECODED FROM PIPE: ", ans)
				if err != nil {
					ans.TaskID = w.taskID
					ans.worker = w
					ans.Error = err.Error()
					// TODO
					log.Print("ERROR: ", err)

				} else {
					ans.worker = w
				}
				w.lastEvent = ans
				w.workerEvent <- &ans

			}
		}
	}()
	err = w.cmd.Start()
	if err != nil {
		log.Print("ERROR: ", err) // TODO
	}
	go func() {
		err := w.cmd.Wait()
		if err != nil {
			switch terr := err.(type) {
			case *exec.ExitError:
				w.workerEvent <- &WorkerStatus{
					worker: w,
					Error:  terr.Error(), // TODO
				}
			default:
				w.workerEvent <- &WorkerStatus{
					worker: w,
					Error:  err.Error(),
				}
			}
		}
	}()

}

// Stop kills the external task
func (w *Worker) Stop() {
	w.cmd.Process.Kill()
	w.commandsPipe.reader.Close()
	w.commandsPipe.writer.Close()
	w.responsesPipe.reader.Close()
	w.responsesPipe.writer.Close()
}

// Call sends a speicifed command to the worker.
// Generally, the args is expected to be JSON-encodable
// data. Konserver does not care about it contents and
// just passes it to the worker.
func (w *Worker) Call(taskID string, fn string, args interface{}) {
	js, err := json.Marshal(workerCall{
		Fn:     fn,
		Args:   args,
		TaskID: taskID,
	})
	if err != nil {
		// TODO
		log.Print("ERROR: ", err)
	}
	w.taskID = taskID
	w.commandsPipe.SendBytes(js)
}

// Reload sends SIGHUP to the running task
func (w *Worker) Reload() {
	w.cmd.Process.Signal(syscall.SIGHUP)
}

func (w *Worker) Info() WorkerInfo {
	ans := WorkerInfo{
		PID:        w.GetPID(),
		TaskID:     w.taskID,
		LastStatus: w.lastEvent.ReadableStatus(),
	}
	return ans
}
