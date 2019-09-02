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
	"os"
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

func TestCollectorConfiguration(t *testing.T) {
	os.Setenv("VESMGR_PRICOLLECTOR_USER", "user123")
	os.Setenv("VESMGR_PRICOLLECTOR_PASSWORD", "pass123")
	os.Setenv("VESMGR_PRICOLLECTOR_PASSPHRASE", "phrase123")
	os.Setenv("VESMGR_PRICOLLECTOR_ADDR", "1.2.3.4")
	os.Setenv("VESMGR_PRICOLLECTOR_PORT", "1234")
	os.Setenv("VESMGR_PRICOLLECTOR_SERVERROOT", "vescollector")
	os.Setenv("VESMGR_PRICOLLECTOR_TOPIC", "sometopic")
	os.Setenv("VESMGR_PRICOLLECTOR_SECURE", "true")

	vesconf := basicVespaConf()
	getCollectorConfiguration(&vesconf)

	assert.Equal(t, "user123", vesconf.PrimaryCollector.User)
	assert.Equal(t, "pass123", vesconf.PrimaryCollector.Password)
	assert.Equal(t, "phrase123", vesconf.PrimaryCollector.PassPhrase)
	assert.Equal(t, "1.2.3.4", vesconf.PrimaryCollector.FQDN)
	assert.Equal(t, 1234, vesconf.PrimaryCollector.Port)
	assert.Equal(t, "vescollector", vesconf.PrimaryCollector.ServerRoot)
	assert.Equal(t, "sometopic", vesconf.PrimaryCollector.Topic)
	assert.Equal(t, true, vesconf.PrimaryCollector.Secure)
}

func TestCollectorConfigurationWhenEnvironmentVariablesAreNotDefined(t *testing.T) {
	os.Unsetenv("VESMGR_PRICOLLECTOR_USER")
	os.Unsetenv("VESMGR_PRICOLLECTOR_PASSWORD")
	os.Unsetenv("VESMGR_PRICOLLECTOR_PASSPHRASE")
	os.Unsetenv("VESMGR_PRICOLLECTOR_ADDR")
	os.Unsetenv("VESMGR_PRICOLLECTOR_PORT")
	os.Unsetenv("VESMGR_PRICOLLECTOR_SERVERROOT")
	os.Unsetenv("VESMGR_PRICOLLECTOR_TOPIC")
	os.Unsetenv("VESMGR_PRICOLLECTOR_SECURE")

	vesconf := basicVespaConf()
	getCollectorConfiguration(&vesconf)

	assert.Equal(t, "", vesconf.PrimaryCollector.User)
	assert.Equal(t, "", vesconf.PrimaryCollector.Password)
	assert.Equal(t, "", vesconf.PrimaryCollector.PassPhrase)
	assert.Equal(t, "", vesconf.PrimaryCollector.FQDN)
	assert.Equal(t, 8443, vesconf.PrimaryCollector.Port)
	assert.Equal(t, "", vesconf.PrimaryCollector.ServerRoot)
	assert.Equal(t, "", vesconf.PrimaryCollector.Topic)
	assert.Equal(t, false, vesconf.PrimaryCollector.Secure)
}

func TestCollectorConfigurationWhenPrimaryCollectorPortIsNotInteger(t *testing.T) {
	os.Setenv("VESMGR_PRICOLLECTOR_PORT", "abcd")
	vesconf := basicVespaConf()
	getCollectorConfiguration(&vesconf)
	assert.Equal(t, 0, vesconf.PrimaryCollector.Port)
}

func TestCollectorConfigurationWhenPrimaryCollectorSecureIsNotTrueOrFalse(t *testing.T) {
	os.Setenv("VESMGR_PRICOLLECTOR_SECURE", "foo")
	vesconf := basicVespaConf()
	getCollectorConfiguration(&vesconf)
	assert.Equal(t, false, vesconf.PrimaryCollector.Secure)
}

func TestYamlGeneration(t *testing.T) {
	buffer := new(bytes.Buffer)
	createVespaConfig(buffer)
	var vesconf VESAgentConfiguration
	err := yaml.Unmarshal(buffer.Bytes(), &vesconf)
	assert.Nil(t, err)
	testBaseConf(t, vesconf)
}
