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
	"testing"
	"time"
	"bytes"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
 )

 func testBaseConf(t *testing.T, vesconf VESAgentConfiguration) {
	assert.Equal(t, "/tmp/data", vesconf.DataDir)
	assert.False(t, vesconf.Debug)
	assert.Equal(t, vesconf.Event.MaxMissed, 2)
	assert.Equal(t, vesconf.Event.RetryInterval, time.Second*5)
	assert.Equal(t, vesconf.Measurement.Prometheus.KeepAlive, time.Second*30)
 }

 func TestBasicConfigContainsCorrectValues(t *testing.T) {
	vesconf := basicVespaConf()
	testBaseConf(t, vesconf)
 }

 func TestYamlGeneratio(t *testing.T) {
	buffer := new(bytes.Buffer)
	createVespaConfig(buffer)
	var vesconf VESAgentConfiguration
	err := yaml.Unmarshal(buffer.Bytes(), &vesconf)
	assert.Nil(t, err)
	testBaseConf(t, vesconf)
 }
