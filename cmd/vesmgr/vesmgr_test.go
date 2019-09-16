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
 */

package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"os"
	"os/exec"
	"testing"
	"time"
)

func init() {
	vesagent.name = "echo" // no need to run real ves-agent
	logger.MdcAdd("Testvesmgr", "0.0.1")
	os.Setenv("VESMGR_HB_INTERVAL", "30s")
	os.Setenv("VESMGR_MEAS_INTERVAL", "30s")
	os.Setenv("VESMGR_PRICOLLECTOR_ADDR", "127.1.1.1")
	os.Setenv("VESMGR_PRICOLLECTOR_PORT", "8443")
	os.Setenv("VESMGR_PROMETHEUS_ADDR", "http://localhost:9090")
}

func TestStartVesagent(t *testing.T) {
	assert.Equal(t, 0, vesagent.Pid)
	ch := make(chan error)
	startVesagent(ch)
	assert.NotEqual(t, 0, vesagent.Pid)
	t.Logf("VES agent pid = %d", vesagent.Pid)
	vesagent.Pid = 0
	err := <-ch
	assert.Nil(t, err)
}

func TestStartVesagentFails(t *testing.T) {
	vesagent.name = "Not-ves-agent"
	assert.Equal(t, 0, vesagent.Pid)
	ch := make(chan error)
	startVesagent(ch)
	err := <-ch
	assert.NotNil(t, err)
	assert.Equal(t, 0, vesagent.Pid)
	vesagent.name = "ves-agent"
}

func TestGetMyIP(t *testing.T) {
	vesmgr.myIPAddress = string("")
	var err error
	vesmgr.myIPAddress, err = getMyIP()
	assert.NotEqual(t, string(""), vesmgr.myIPAddress)
	assert.Equal(t, nil, err)
}

func TestMainLoopSupervision(t *testing.T) {
	chXAppNotifications := make(chan []byte)
	chSupervision := make(chan chan string)
	chVesagent := make(chan error)
	go runVesmgr(chVesagent, chSupervision, chXAppNotifications)

	ch := make(chan string)
	chSupervision <- ch
	reply := <-ch
	assert.Equal(t, "OK", reply)
}

func TestMainLoopVesagentError(t *testing.T) {
	if os.Getenv("TEST_VESPA_EXIT") == "1" {
		// we're run in a new process, now make vesmgr main loop exit
		chXAppNotifications := make(chan []byte)
		chSupervision := make(chan chan string)
		chVesagent := make(chan error)
		go runVesmgr(chVesagent, chSupervision, chXAppNotifications)

		chVesagent <- errors.New("vesagent killed")
		// we should never actually end up to this sleep, since the runVesmgr should exit
		time.Sleep(3 * time.Second)
		return
	}

	// Run the vesmgr exit test as a separate process
	cmd := exec.Command(os.Args[0], "-test.run=TestMainLoopVesagentError")
	cmd.Env = append(os.Environ(), "TEST_VESPA_EXIT=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	// check that vesmgr existed with status 1
	e, ok := err.(*exec.ExitError)
	assert.Equal(t, true, ok)
	assert.Equal(t, "exit status 1", e.Error())
}
