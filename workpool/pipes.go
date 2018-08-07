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
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os/exec"
)

const (
	initialBufferSize = 16384
)

// CommandPipe is used to send
// text commands (typically it is JSON)
// to a worker.
type CommandPipe struct {
	writer *io.PipeWriter
	reader *io.PipeReader
}

// NewCommandPipe is the default factory method
// for CommandPipe
func NewCommandPipe() *CommandPipe {
	ans := &CommandPipe{}
	ans.reader, ans.writer = io.Pipe()
	return ans
}

// Register connects the pipe with
// a provided command (Cmd)
func (cp *CommandPipe) Register(cmd *exec.Cmd) {
	cmd.Stdin = cp.reader
}

// SendBytes sends specified bytes to the pipe
func (cp *CommandPipe) SendBytes(command []byte) {
	_, err := cp.writer.Write(append(command, '\n'))
	if err != nil {
		log.Print("ERROR: ", err) // TODO
	}
}

// ---------------------------------------------------------------

// ResponsePipe is used to receive data from worker
// command line program. As it is expected that data
// can be quite large, an internal Scanner can be
// configured to work with a buffer of a custom size.
type ResponsePipe struct {
	writer        *io.PipeWriter
	reader        *io.PipeReader
	rChan         chan string
	maxBufferSize int
}

// NewResponsePipe is a default factory for ResponsePipe
func NewResponsePipe(maxBufferSize int) *ResponsePipe {
	ans := &ResponsePipe{
		maxBufferSize: maxBufferSize,
	}
	ans.rChan = make(chan string)
	ans.reader, ans.writer = io.Pipe()
	return ans
}

// Register configures the pipe to work
// with a specified command (Cmd).
func (cp *ResponsePipe) Register(cmd *exec.Cmd) {
	cmd.Stdout = cp.writer
	go func() {
		sc := bufio.NewScanner(cp.reader)
		sc.Buffer(make([]byte, initialBufferSize), cp.maxBufferSize)
		for sc.Scan() {
			cp.rChan <- sc.Text()
		}
		err := sc.Err()
		if err != nil {
			log.Print("ERROR: Scanner error - ", err)
			ans := make(map[string]string)
			ans["error"] = err.Error()
			jsonAns, err := json.Marshal(ans)
			if err != nil {
				log.Print("ERROR: Broken Scanner error handling: ", err)
				cp.rChan <- "\n" // we must end the command properly otherwise it gets stuck

			} else {
				cp.rChan <- string(jsonAns)
			}
		}
	}()

}

// Channel returns pipes channel where
// received text lines are sent.
func (cp *ResponsePipe) Channel() chan string {
	return cp.rChan
}
