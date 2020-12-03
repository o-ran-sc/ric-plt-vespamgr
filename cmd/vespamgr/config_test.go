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
 */
package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var vespaMgr *VespaMgr

func testBaseConf(t *testing.T, vesconf VESAgentConfiguration) {
	vespaMgr = NewVespaMgr()

	assert.Equal(t, "/tmp/data", vesconf.DataDir)
	assert.False(t, vesconf.Debug)
	assert.Equal(t, vesconf.Event.MaxMissed, 2)
	assert.Equal(t, vesconf.Event.RetryInterval, time.Second*5)
	assert.Equal(t, vesconf.Measurement.Prometheus.KeepAlive, time.Second*30)
	assert.Equal(t, vesconf.Event.VNFName, defaultVNFName)
	assert.Equal(t, vesconf.Event.NfNamingCode, defaultNFNamingCode)
	assert.Equal(t, vesconf.Event.ReportingEntityName, "Vespa")
	// depending on the credentials with which this test is run,
	// root or non-root, the code either reads the UUID from the file or
	// ends up using the default id. Just check the length here,
	// not the actual value.
	assert.Len(t, vesconf.Event.ReportingEntityID, len(defaultReportingEntityID))
}

func TestBasicConfigContainsCorrectValues(t *testing.T) {
	vesconf := vespaMgr.BasicVespaConf()
	testBaseConf(t, vesconf)
}

func TestYamlGenerationWithoutXAppsConfig(t *testing.T) {
	buffer := new(bytes.Buffer)
	vespaMgr.CreateConfig(buffer, []byte{})
	var vesconf VESAgentConfiguration
	err := yaml.Unmarshal(buffer.Bytes(), &vesconf)
	assert.Nil(t, err)
	testBaseConf(t, vesconf)
	assert.Empty(t, vesconf.Measurement.Prometheus.Rules.Metrics)
}

func TestYamlGenerationWithXAppsConfig(t *testing.T) {
	buffer := new(bytes.Buffer)
	bytes, err := ioutil.ReadFile("../../test/xApp_config_test_output.json")
	assert.Nil(t, err)
	vespaMgr.CreateConfig(buffer, bytes)
	var vesconf VESAgentConfiguration
	err = yaml.Unmarshal(buffer.Bytes(), &vesconf)
	assert.Nil(t, err)
	testBaseConf(t, vesconf)
	assert.Len(t, vesconf.Measurement.Prometheus.Rules.Metrics, 4)
}

// Helper function for the metrics parsing tests
func metricsStringToInterfaceArray(metrics string) []interface{} {
	var metricsArray map[string][]interface{}
	json.Unmarshal([]byte(metrics), &metricsArray)
	return metricsArray["metrics"]
}

func TestParseMetricsRules(t *testing.T) {
	metricsJSON := `{"metrics": [
			{ "name": "ricxapp_RMR_Received", "objectName": "ricxappRMRreceivedCounter", "objectInstance": "ricxappRMRReceived", "counterId": "0011" },
			{ "name": "ricxapp_RMR_ReceiveError", "objectName": "ricxappRMRReceiveErrorCounter", "objectInstance": "ricxappRMRReceiveError", "counterId": "0011" },
			{ "name": "ricxapp_RMR_Transmitted", "objectName": "ricxappRMRTransmittedCounter", "objectInstance": "ricxappRMRTransmitted", "counterId": "0011" },
			{ "name": "ricxapp_RMR_TransmitError", "objectName": "ricxappRMRTransmitErrorCounter", "objectInstance": "ricxappRMRTransmitError", "counterId": "0011" },
			{ "name": "ricxapp_SDL_Stored", "objectName": "ricxappSDLStoredCounter", "objectInstance": "ricxappSDLStored", "counterId": "0011" },
			{ "name": "ricxapp_SDL_StoreError", "objectName": "ricxappSDLStoreErrorCounter", "objectInstance": "ricxappSDLStoreError", "counterId": "0011" } ]}`
	appMetrics := make(AppMetrics)
	m := metricsStringToInterfaceArray(metricsJSON)
	appMetrics = vespaMgr.ParseMetricsRules(m, appMetrics, "SEP/XAPP", "X2", "1234", "60")
	assert.Len(t, appMetrics, 6)
	assert.Equal(t, "ricxappRMRreceivedCounter", appMetrics["ricxapp_RMR_Received"].ObjectName)
	assert.Equal(t, "ricxappRMRTransmitErrorCounter", appMetrics["ricxapp_RMR_TransmitError"].ObjectName)
	assert.Equal(t, "ricxappSDLStoreError", appMetrics["ricxapp_SDL_StoreError"].ObjectInstance)
}

func TestParseMetricsRulesNoMetrics(t *testing.T) {
	appMetrics := make(AppMetrics)
	metricsJSON := `{"metrics": []`
	m := metricsStringToInterfaceArray(metricsJSON)
	appMetrics = vespaMgr.ParseMetricsRules(m, appMetrics, "SEP/XAPP", "X2", "1234", "60")
	assert.Empty(t, appMetrics)
}

func TestParseMetricsRulesAdditionalFields(t *testing.T) {
	appMetrics := make(AppMetrics)
	metricsJSON := `{"metrics": [
			{ "additionalField": "valueIgnored", "name": "ricxapp_RMR_Received", "objectName": "ricxappRMRreceivedCounter", "objectInstance": "ricxappRMRReceived", "counterId": "0011" }]}`
	m := metricsStringToInterfaceArray(metricsJSON)
	appMetrics = vespaMgr.ParseMetricsRules(m, appMetrics, "SEP/XAPP", "X2", "1234", "60")
	assert.Len(t, appMetrics, 1)
	assert.Equal(t, "ricxappRMRreceivedCounter", appMetrics["ricxapp_RMR_Received"].ObjectName)
	assert.Equal(t, "ricxappRMRReceived", appMetrics["ricxapp_RMR_Received"].ObjectInstance)
}

func TestParseMetricsRulesMissingFields(t *testing.T) {
	appMetrics := make(AppMetrics)
	metricsJSON := `{"metrics": [
			{ "name": "ricxapp_RMR_Received", "objectName": "ricxappRMRreceivedCounter", "objectInstance": "ricxappRMRReceived", "counterId": "0011" },
			{ "name": "ricxapp_RMR_ReceiveError", "objectInstance": "ricxappRMRReceiveError" },
			{ "name": "ricxapp_RMR_Transmitted", "objectName": "ricxappRMRTransmittedCounter", "objectInstance": "ricxappRMRTransmitted", "counterId": "0011" }]}`
	m := metricsStringToInterfaceArray(metricsJSON)
	appMetrics = vespaMgr.ParseMetricsRules(m, appMetrics, "SEP/XAPP", "X2", "1234", "60")
	assert.Len(t, appMetrics, 2)
	assert.Equal(t, "ricxappRMRreceivedCounter", appMetrics["ricxapp_RMR_Received"].ObjectName)
	assert.Equal(t, "ricxappRMRTransmittedCounter", appMetrics["ricxapp_RMR_Transmitted"].ObjectName)
	_, ok := appMetrics["ricxapp_RMR_ReceiveError"]
	assert.False(t, ok)
}

func TestParseMetricsRulesDuplicateDefinitionIsIgnored(t *testing.T) {
	appMetrics := make(AppMetrics)
	metricsJSON := `{"metrics": [
			{ "name": "ricxapp_RMR_Received", "objectName": "ricxappRMRreceivedCounter", "objectInstance": "ricxappRMRReceived", "counterId": "0011" },
			{ "name": "ricxapp_RMR_Received", "objectName": "ricxappRMRreceivedCounterXXX", "objectInstance": "ricxappRMRReceivedXXX", "counterId": "0011" },
			{ "name": "ricxapp_RMR_Transmitted", "objectName": "ricxappRMRTransmittedCounter", "objectInstance": "ricxappRMRTransmitted", "counterId": "0011" }]}`
	m := metricsStringToInterfaceArray(metricsJSON)
	appMetrics = vespaMgr.ParseMetricsRules(m, appMetrics, "SEP/XAPP", "X2", "1234", "60")
	assert.Len(t, appMetrics, 2)
	assert.Equal(t, "ricxappRMRreceivedCounter", appMetrics["ricxapp_RMR_Received"].ObjectName)
	assert.Equal(t, "ricxappRMRReceived", appMetrics["ricxapp_RMR_Received"].ObjectInstance)
}

func TestParseMetricsRulesIncrementalFillOfAppMetrics(t *testing.T) {
	appMetrics := make(AppMetrics)
	metricsJSON1 := `{"metrics": [
			{ "name": "ricxapp_RMR_Received", "objectName": "ricxappRMRreceivedCounter", "objectInstance": "ricxappRMRReceived", "counterId": "0011" }]}`
	metricsJSON2 := `{"metrics": [
			{ "name": "ricxapp_RMR_Transmitted", "objectName": "ricxappRMRTransmittedCounter", "objectInstance": "ricxappRMRTransmitted", "counterId": "0011" }]}`
	m1 := metricsStringToInterfaceArray(metricsJSON1)
	m2 := metricsStringToInterfaceArray(metricsJSON2)
	appMetrics = vespaMgr.ParseMetricsRules(m1, appMetrics, "SEP/XAPP", "X2", "1234", "60")
	appMetrics = vespaMgr.ParseMetricsRules(m2, appMetrics, "SEP/XAPP", "X2", "1234", "60")
	assert.Len(t, appMetrics, 2)
	assert.Equal(t, "ricxappRMRreceivedCounter", appMetrics["ricxapp_RMR_Received"].ObjectName)
	assert.Equal(t, "ricxappRMRReceived", appMetrics["ricxapp_RMR_Received"].ObjectInstance)
}

func TestParseXAppDescriptor(t *testing.T) {
	appMetrics := make(AppMetrics)
	bytes, err := ioutil.ReadFile("../../test/xApp_config_test_output.json")
	assert.Nil(t, err)

	appMetrics = vespaMgr.ParseMetricsFromDescriptor(bytes, appMetrics)
	assert.Len(t, appMetrics, 4)
	assert.Equal(t, "App1ExampleCounterOneObject", appMetrics["App1ExampleCounterOne"].ObjectName)
	assert.Equal(t, "App1ExampleCounterOneObjectInstance", appMetrics["App1ExampleCounterOne"].ObjectInstance)
	assert.Equal(t, "App1ExampleCounterTwoObject", appMetrics["App1ExampleCounterTwo"].ObjectName)
	assert.Equal(t, "App1ExampleCounterTwoObjectInstance", appMetrics["App1ExampleCounterTwo"].ObjectInstance)
	assert.Equal(t, "App2ExampleCounterOneObject", appMetrics["App2ExampleCounterOne"].ObjectName)
	assert.Equal(t, "App2ExampleCounterOneObjectInstance", appMetrics["App2ExampleCounterOne"].ObjectInstance)
	assert.Equal(t, "App2ExampleCounterTwoObject", appMetrics["App2ExampleCounterTwo"].ObjectName)
	assert.Equal(t, "App2ExampleCounterTwoObjectInstance", appMetrics["App2ExampleCounterTwo"].ObjectInstance)
}

func TestParseXAppDescriptorWithNoConfig(t *testing.T) {
	metricsJSON := `[{{"metadata": "something", "descriptor": "somethingelse"}},
	                 {{"metadata": "something", "descriptor": "somethingelse"}}]`
	metricsBytes := []byte(metricsJSON)
	appMetrics := make(AppMetrics)
	appMetrics = vespaMgr.ParseMetricsFromDescriptor(metricsBytes, appMetrics)
	assert.Empty(t, appMetrics)
}

func TestParseXAppDescriptorWithNoMetrics(t *testing.T) {
	metricsJSON := `[{{"metadata": "something", "descriptor": "somethingelse", "config":{}},
	                 {{"metadata": "something", "descriptor": "somethingelse", "config":{}}}]`
	metricsBytes := []byte(metricsJSON)
	appMetrics := make(AppMetrics)
	appMetrics = vespaMgr.ParseMetricsFromDescriptor(metricsBytes, appMetrics)
	assert.Empty(t, appMetrics)
}
