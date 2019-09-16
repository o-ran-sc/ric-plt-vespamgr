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
	"encoding/json"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"strconv"
	"time"
)

func basicVespaConf() VESAgentConfiguration {
	var vespaconf = VESAgentConfiguration{
		DataDir: "/tmp/data",
		Debug:   false,
		Event: EventConfiguration{
			VNFName:           "vespa-demo",                          // XXX
			ReportingEntityID: "1af5bfa9-40b4-4522-b045-40e54f0310f", // XXX
			MaxSize:           2000000,
			NfNamingCode:      "hsxp",
			NfcNamingCodes: []NfcNamingCode{
				NfcNamingCode{
					Type:  "oam",
					Vnfcs: []string{"lr-ope-0", "lr-ope-1", "lr-ope-2"},
				},
				NfcNamingCode{
					Type:  "etl",
					Vnfcs: []string{"lr-pro-0", "lr-pro-1"},
				},
			},
			RetryInterval: time.Second * 5,
			MaxMissed:     2,
		},
		Measurement: MeasurementConfiguration{
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

type AppMetricsStruct struct {
	ObjectName     string
	ObjectInstance string
	// xxx add labels here
}

type AppMetrics map[string]AppMetricsStruct

// Parses the metrics data from an array of bytes, which is expected to contain a JSON
// array with structs of the following format:
//
// { ...
//   "config" : {
//     "metrics": [
//       { "name": "...", "objectName": "...", "objectInstamce": "..." },
//       ...
//     ]
//   }
// }
func parseMetricsFromXAppDescriptor(descriptor []byte, appMetrics AppMetrics) AppMetrics {
	var desc []map[string]interface{}
	json.Unmarshal(descriptor, &desc)

	for _, app := range desc {
		config, config_ok := app["config"]
		if config_ok {
			metrics, metrics_ok := config.(map[string]interface{})["metrics"]
			if metrics_ok {
				parseMetricsRules(metrics.([]interface{}), appMetrics)
			}
		}
	}
	return appMetrics
}

// Parses the metrics data from an array of interfaces, which are expected to be maps
// of the following format:
//    { "name": xxx, "objectName": yyy, "objectInstance": zzz }
// Entries, which do not have all the necessary fields, are ignored.
func parseMetricsRules(metricsMap []interface{}, appMetrics AppMetrics) AppMetrics {
	for _, element := range metricsMap {
		name, name_ok := element.(map[string]interface{})["name"].(string)
		if name_ok {
			_, already_found := appMetrics[name]
			objectName, objectName_ok := element.(map[string]interface{})["objectName"].(string)
			objectInstance, objectInstance_ok := element.(map[string]interface{})["objectInstance"].(string)
			if !already_found && objectName_ok && objectInstance_ok {
				appMetrics[name] = AppMetricsStruct{objectName, objectInstance}
				logger.Info("parsed counter %s %s %s", name, objectName, objectInstance)
			}
			if already_found {
				logger.Info("skipped duplicate counter %s", name)
			}
		}
	}
	return appMetrics
}

func getRules(vespaconf *VESAgentConfiguration, xAppConfig []byte) {
	appMetrics := make(AppMetrics)
	parseMetricsFromXAppDescriptor(xAppConfig, appMetrics)

	makeRule := func(expr string, obj_name string, obj_instance string) MetricRule {
		return MetricRule{
			Target:         "AdditionalObjects",
			Expr:           expr,
			ObjectInstance: obj_instance,
			ObjectName:     obj_name,
			ObjectKeys: []Label{
				Label{
					Name: "ricComponentName",
					Expr: "'{{.labels.kubernetes_name}}'",
				},
			},
		}
	}
	var metricsMap map[string][]interface{}
	json.Unmarshal(xAppConfig, &metricsMap)
	metrics := parseMetricsRules(metricsMap["metrics"], appMetrics)

	vespaconf.Measurement.Prometheus.Rules.Metrics = make([]MetricRule, 0, len(metrics))
	for key, value := range metrics {
		vespaconf.Measurement.Prometheus.Rules.Metrics = append(vespaconf.Measurement.Prometheus.Rules.Metrics, makeRule(key, value.ObjectName, value.ObjectInstance))
	}
	if len(vespaconf.Measurement.Prometheus.Rules.Metrics) == 0 {
		logger.Info("vespa config with empty metrics")
	}
}

func getCollectorConfiguration(vespaconf *VESAgentConfiguration) {
	vespaconf.PrimaryCollector.User = os.Getenv("VESMGR_PRICOLLECTOR_USER")
	vespaconf.PrimaryCollector.Password = os.Getenv("VESMGR_PRICOLLECTOR_PASSWORD")
	vespaconf.PrimaryCollector.PassPhrase = os.Getenv("VESMGR_PRICOLLECTOR_PASSPHRASE")
	vespaconf.PrimaryCollector.FQDN = os.Getenv("VESMGR_PRICOLLECTOR_ADDR")
	vespaconf.PrimaryCollector.ServerRoot = os.Getenv("VESMGR_PRICOLLECTOR_SERVERROOT")
	vespaconf.PrimaryCollector.Topic = os.Getenv("VESMGR_PRICOLLECTOR_TOPIC")
	port_str := os.Getenv("VESMGR_PRICOLLECTOR_PORT")
	if port_str == "" {
		vespaconf.PrimaryCollector.Port = 8443
	} else {
		port, _ := strconv.Atoi(port_str)
		vespaconf.PrimaryCollector.Port = port
	}
	secure_str := os.Getenv("VESMGR_PRICOLLECTOR_SECURE")
	if secure_str == "true" {
		vespaconf.PrimaryCollector.Secure = true
	} else {
		vespaconf.PrimaryCollector.Secure = false
	}
}

func createVespaConfig(writer io.Writer, xAppStatus []byte) {
	vespaconf := basicVespaConf()
	getRules(&vespaconf, xAppStatus)
	getCollectorConfiguration(&vespaconf)
	err := yaml.NewEncoder(writer).Encode(vespaconf)
	if err != nil {
		logger.Error("Cannot write vespa conf file: %s", err.Error())
		return
	}
}
