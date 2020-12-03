/*
 *  Copyright (c) 2019 AT&T Intellectual Property.
 *  Copyright (c) 2018-2019 Nokia.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 *  This source code is part of the near-RT RIC (RAN Intelligent Controller)
 *  platform project (RICP).
 *
 */
package main

import (
	"os"
	"os/exec"
)

type CommandRunner struct {
	exe  string
	args []string
	cmd  *exec.Cmd
}

func (r *CommandRunner) Run(result chan error) {
	r.cmd = exec.Command(r.exe, r.args...)
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr
	err := r.cmd.Start()
	go func() {
		if err != nil {
			result <- err
		} else {
			result <- r.cmd.Wait()
		}
	}()
}

func (r *CommandRunner) Kill() error {
	return r.cmd.Process.Kill()
}

func NewCommandRunner(exe string, arg ...string) *CommandRunner {
	r := &CommandRunner{exe: exe, args: arg}
	return r
}
