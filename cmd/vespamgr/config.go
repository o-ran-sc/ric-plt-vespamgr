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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	app "gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"gopkg.in/yaml.v2"
)

const defaultReportingEntityID = "00000000-0000-0000-0000-000000000000"
const defaultVNFName = "Vespa"
const defaultNFNamingCode = "ricp"

func (v *VespaMgr) readSystemUUID() string {
	data, err := ioutil.ReadFile("/sys/class/dmi/id/product_uuid")
	if err != nil {
		return defaultReportingEntityID
	}
	return strings.TrimSpace(string(data))
}

func (v *VespaMgr) getVNFName() string {
	VNFName := os.Getenv("VESMGR_VNFNAME")
	if VNFName == "" {
		return defaultVNFName
	}
	return VNFName
}

func (v *VespaMgr) getNFNamingCode() string {
	NFNamingCode := os.Getenv("VESMGR_NFNAMINGCODE")
	if NFNamingCode == "" {
		return defaultNFNamingCode
	}
	return NFNamingCode
}

func (v *VespaMgr) BasicVespaConf() VESAgentConfiguration {
	var vespaconf = VESAgentConfiguration{
		DataDir: "/tmp/data",
		Debug:   false,
		Event: EventConfiguration{
			VNFName:             v.getVNFName(),
			ReportingEntityName: "Vespa",
			ReportingEntityID:   v.readSystemUUID(),
			MaxSize:             100000000,
			NfNamingCode:        v.getNFNamingCode(),
			NfcNamingCodes:      []NfcNamingCode{},
			RetryInterval:       time.Second * 5,
			MaxMissed:           2,
		},
		Measurement: MeasurementConfiguration{
			// Domain abbreviation has to be set to “Mvfs” for VES 5.3,
			// and to “Measurement” for later VES interface versions.
			DomainAbbreviation:   "Mvfs",
			MaxBufferingDuration: time.Hour,
			Prometheus: PrometheusConfig{
				Timeout:   time.Second * 30,
				KeepAlive: time.Second * 30,
				Rules: MetricRules{
					DefaultValues: &MetricRule{
						VMIDLabel: "'{{.labels.instance}}'",
					},
				},
			},
		},
	}
	return vespaconf
}

// Parses the metrics data from an array of bytes, which is expected to contain a JSON
// array with structs of the following format:
//
// { ...
//   "config" : {
//	   "measurements": [
//      {
//     	  "metrics": [
//          { "name": "...", "objectName": "...", "objectInstamce": "..." },
//           ...
//         ]
//       }
//       ...
//     ]
//    }
// }
func (v *VespaMgr) ParseMetricsFromDescriptor(descriptor []byte, appMetrics AppMetrics) AppMetrics {
	var desc []map[string]interface{}
	json.Unmarshal(descriptor, &desc)

	for _, appl := range desc {
		config, configOk := appl["config"]
		if !configOk {
			app.Logger.Info("No xApp config found!")
			continue
		}
		measurements, measurementsOk := config.(map[string]interface{})["measurements"]
		if !measurementsOk {
			app.Logger.Info("No xApp metrics found!")
			continue
		}

		for _, m := range measurements.([]interface{}) {
			moId, moIdOk := m.(map[string]interface{})["moId"].(string)
			measType, measTypeOk := m.(map[string]interface{})["measType"].(string)
			measId, measIdOk := m.(map[string]interface{})["measId"].(string)
			measInterval, measIntervalOk := m.(map[string]interface{})["measInterval"].(string)
			metrics, metricsOk := m.(map[string]interface{})["metrics"]
			if !metricsOk || !measTypeOk || !measIdOk || !moIdOk || !measIntervalOk {
				app.Logger.Info("No metrics found for moId=%s measType=%s measId=%s measInterval=%s", moId, measId, measType, measInterval)
				continue
			}
			app.Logger.Info("Parsed measurement: moId=%s type=%s id=%s interval=%s", moId, measType, measId, measInterval)

			v.ParseMetricsRules(metrics.([]interface{}), appMetrics, moId, measType, measId, measInterval)
		}
	}
	return appMetrics
}

// Parses the metrics data from an array of interfaces, which are expected to be maps
// of the following format:
//    { "name": xxx, "objectName": yyy, "objectInstance": zzz }
// Entries, which do not have all the necessary fields, are ignored.
func (v *VespaMgr) ParseMetricsRules(metricsMap []interface{}, appMetrics AppMetrics, moId, measType, measId, measInterval string) AppMetrics {
	for _, element := range metricsMap {
		name, nameOk := element.(map[string]interface{})["name"].(string)
		if nameOk {
			_, alreadyFound := appMetrics[name]
			objectName, objectNameOk := element.(map[string]interface{})["objectName"].(string)
			objectInstance, objectInstanceOk := element.(map[string]interface{})["objectInstance"].(string)
			counterId, counterIdOk := element.(map[string]interface{})["counterId"].(string)
			if !alreadyFound && objectNameOk && objectInstanceOk && counterIdOk {
				appMetrics[name] = AppMetricsStruct{moId, measType, measId, measInterval, objectName, objectInstance, counterId}
				app.Logger.Info("Parsed counter name=%s %s/%s  M%sC%s", name, objectName, objectInstance, measId, counterId)
			}
			if alreadyFound {
				app.Logger.Info("skipped duplicate counter %s", name)
			}
		}
	}
	return appMetrics
}

func (v *VespaMgr) GetRules(vespaconf *VESAgentConfiguration, xAppConfig []byte) bool {
	makeRule := func(expr string, value AppMetricsStruct) MetricRule {
		return MetricRule{
			Target:         "AdditionalObjects",
			Expr:           expr,
			ObjectInstance: fmt.Sprintf("%s:%s", value.ObjectInstance, value.CounterId),
			ObjectName:     value.ObjectName,
			ObjectKeys: []Label{
				{Name: "ricComponentName", Expr: "'{{.labels.kubernetes_name}}'"},
				{Name: "moId", Expr: value.MoId},
				{Name: "measType", Expr: value.MeasType},
				{Name: "measId", Expr: value.MeasId},
				{Name: "measInterval", Expr: value.MeasInterval},
			},
		}
	}
	appMetrics := make(AppMetrics)
	metrics := v.ParseMetricsFromDescriptor(xAppConfig, appMetrics)

	if v.pltFileCreated {
		pltConfig, err := ioutil.ReadFile(app.Config.GetString("controls.pltFile"))
		if err != nil {
			app.Logger.Error("Unable to read platform config file: %v", err)
		} else {
			metrics = v.ParseMetricsFromDescriptor(pltConfig, metrics)
		}
	}
    
	// Adding Platform Counters
	pltCounterFile :=  app.Config.GetString("controls.pltCounterFile")
	bytes, err := ioutil.ReadFile(pltCounterFile)
	if err != nil{
		app.Logger.Error("Platform Matrices Configuration File not found")
	} else {

		metrics = v.ParseMetricsFromDescriptor(bytes,metrics)
	}
	

	vespaconf.Measurement.Prometheus.Rules.Metrics = make([]MetricRule, 0, len(metrics))
	for key, value := range metrics {
		vespaconf.Measurement.Prometheus.Rules.Metrics = append(vespaconf.Measurement.Prometheus.Rules.Metrics, makeRule(key, value))
	}
	if len(vespaconf.Measurement.Prometheus.Rules.Metrics) == 0 {
		app.Logger.Info("vespa config with empty metrics")
	}

	return len(vespaconf.Measurement.Prometheus.Rules.Metrics) > 0
}

func (v *VespaMgr) GetCollectorConfiguration(vespaconf *VESAgentConfiguration) {
	vespaconf.PrimaryCollector.User = app.Config.GetString("controls.collector.primaryUser")
	vespaconf.PrimaryCollector.Password = app.Config.GetString("controls.collector.primaryPassword")
	vespaconf.PrimaryCollector.PassPhrase = ""
	vespaconf.PrimaryCollector.FQDN = app.Config.GetString("controls.collector.primaryAddr")
	vespaconf.PrimaryCollector.ServerRoot = app.Config.GetString("controls.collector.serverRoot")
	vespaconf.PrimaryCollector.Topic = ""
	vespaconf.PrimaryCollector.Port = app.Config.GetInt("controls.collector.primaryPort")
	vespaconf.PrimaryCollector.Secure = app.Config.GetBool("controls.collector.secure")
}

func (v *VespaMgr) CreateConfig(writer io.Writer, xAppStatus []byte) {

	
	vespaconf := v.BasicVespaConf()
	v.GetRules(&vespaconf, xAppStatus)
	v.GetCollectorConfiguration(&vespaconf)
    
	err := yaml.NewEncoder(writer).Encode(vespaconf)
	if err != nil {
		app.Logger.Error("Cannot write vespa conf file: %s", err.Error())
		return
	}
	app.Logger.Info("Config file written to: %s", app.Config.GetString("controls.vesagent.configFile"))
}

