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
)

func basicVespaConf() VESAgentConfiguration {
	var vespaconf = VESAgentConfiguration {
		DataDir: "/tmp/data",
		Debug:   false,
		PrimaryCollector: CollectorConfiguration {
			User: "user",
			Password: "pass",
			PassPhrase: "pass",
		},
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
			Target: "AdditionalObject",
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

func createVespaConfig(writer io.Writer) {
	vespaconf := basicVespaConf()
	getRules(&vespaconf)
	err := yaml.NewEncoder(writer).Encode(vespaconf)
	if err != nil {
		logger.Error("Cannot write vespa conf file: %s", err.Error())
		return
	}
}