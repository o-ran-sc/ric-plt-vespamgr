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
	"gopkg.in/yaml.v2"
	"time"
	"io"
	"os"
	"strconv"
)

func basicVespaConf() VESAgentConfiguration {
	var vespaconf = VESAgentConfiguration {
		DataDir: "/tmp/data",
		Debug:   false,
		Event: EventConfiguration {
			VNFName: "vespa-demo", // XXX
			ReportingEntityID: "1af5bfa9-40b4-4522-b045-40e54f0310f", // XXX
			MaxSize: 2000000,
			NfNamingCode: "hsxp",
			NfcNamingCodes: [] NfcNamingCode {
				NfcNamingCode {
					Type: "oam",
					Vnfcs: [] string {"lr-ope-0","lr-ope-1","lr-ope-2"},
				},
				NfcNamingCode {
					Type: "etl",
					Vnfcs: [] string {"lr-pro-0","lr-pro-1"},
				},
			},
			RetryInterval: time.Second * 5,
			MaxMissed: 2,
		},
		Measurement: MeasurementConfiguration {
			DomainAbbreviation: "Mvfs",
			MaxBufferingDuration: time.Hour,
			Prometheus: PrometheusConfig {
				Timeout: time.Second * 30,
				KeepAlive: time.Second * 30,
				Rules: MetricRules {
					DefaultValues: &MetricRule {
						VMIDLabel: "'{{.labels.instance}}'",
					},
				},
			},
		},
	}
	return vespaconf
}

func getRules(vespaconf *VESAgentConfiguration) {
	// XXX
	makeRule := func(expr string, obj_name string, obj_instance string) MetricRule {
		return MetricRule {
			Target: "AdditionalObjects",
			Expr: expr,
			ObjectInstance: obj_instance,
			ObjectName: obj_name,
			ObjectKeys: [] Label {
				Label {
					Name: "ricComponentName",
					Expr: "'{{.labels.app_kubernetes_io_instance}}'",
				},
			},
		}
	}
	// Hard coded for now
	vespaconf.Measurement.Prometheus.Rules.Metrics = []MetricRule {
		makeRule("ricxapp_RMR_Received", "ricxappRMRreceivedCounter", "ricxappRMRReceived"),
		makeRule("ricxapp_RMR_ReceiveError", "ricxappRMRReceiveErrorCounter", "ricxappRMRReceiveError"),
		makeRule("ricxapp_RMR_Transmitted", "ricxappRMRTransmittedCounter", "ricxappRMRTransmitted"),
		makeRule("ricxapp_RMR_TransmitError", "ricxappRMRTransmitErrorCounter", "ricxappRMRTransmitError"),
		makeRule("ricxapp_SDL_Stored", "ricxappSDLStoredCounter", "ricxappSDLStored"),
		makeRule("ricxapp_SDL_StoreError", "ricxappSDLStoreErrorCounter", "ricxappSDLStoreError"),
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

func createVespaConfig(writer io.Writer) {
	vespaconf := basicVespaConf()
	getRules(&vespaconf)
	getCollectorConfiguration(&vespaconf)
	err := yaml.NewEncoder(writer).Encode(vespaconf)
	if err != nil {
		logger.Error("Cannot write vespa conf file: %s", err.Error())
		return
	}
}
