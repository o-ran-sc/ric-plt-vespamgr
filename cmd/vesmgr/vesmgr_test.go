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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
	ch := startVesagent()
	assert.NotEqual(t, 0, vesagent.Pid)
	t.Logf("VES agent pid = %d", vesagent.Pid)
	vesagent.Pid = 0
	err := <-ch
	assert.Nil(t, err)
}

func TestStartVesagentFails(t *testing.T) {

	vesagent.name = "Not-ves-agent"
	assert.Equal(t, 0, vesagent.Pid)
	ch := startVesagent()
	err := <-ch
	assert.NotNil(t, err)
	assert.Equal(t, 0, vesagent.Pid)
	vesagent.name = "ves-agent"
}
