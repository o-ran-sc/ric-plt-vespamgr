..
.. Copyright (c) 2019 AT&T Intellectual Property.
..
.. Copyright (c) 2019 Nokia.
..
..
.. Licensed under the Creative Commons Attribution 4.0 International
..
.. Public License (the "License"); you may not use this file except
..
.. in compliance with the License. You may obtain a copy of the License at
..
..
..     https://creativecommons.org/licenses/by/4.0/
..
..
.. Unless required by applicable law or agreed to in writing, documentation
..
.. distributed under the License is distributed on an "AS IS" BASIS,
..
.. WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
..
.. See the License for the specific language governing permissions and
..
.. limitations under the License.
..


VESPA Manager Overview
======================

The VESPA Manager uses the VES Agent (https://github.com/nokia/ONAP-VESPA) to adapt near-RT RIC internal statistics' collection using Prometheus to scrape metrics from platform and xApp microservices and forward to ONAP or VES Collector via VES interface (VNF event streaming).

The VESPA Manager deployment runs two processes: the VESPA manager and the VES Agent (i.s. VESPA). The VESPA manager starts and configures the VES Agent. 
The VES Agent is a service acting as a bridge between Prometheus and ONAP / VES Collector.

Application Metrics Definition
==============================

The application metrics are defined in the application descriptor. For each counter, the following fields are required in the "metrics" section of the descriptor:

* name - Prometheus name of the counter
* objectName - object name in VES
* objectInstance - object instance in VE

The VESPA manager receives the application metrics configuration from the application manager. It subscribes the app notification messages from the application manager, and after having received one, requests the latest application configuration, creates the VES Agent configuration based on it,
and restarts the VES Agent.

The VES Agent does not report any other metrics to VES.

Prometheus Configuration
========================

The VES Agent reads the ricComponentName from Prometheus label
"kubernetes_name".

VES Collector Event Format
========================

The VES Agent transmits events to the VES Collector in the VES Common Event Format v5.4.1. The Common Event Format is expressed in JSON schema v28.4.1.

VES Event Listener 5.4.1:
<https://docs.onap.org/en/casablanca/submodules/vnfsdk/model.git/docs/files/VESEventListener.html>

JSON schema v28.4.1:
<https://github.com/nokia/ONAP-VESPA/blob/8e9d9e93bb00bed0f5402c9de9502385d5e80acc/doc/CommonEventFormat_28.4.1.json>

Environment Variables
========================

The VESPA manager container requires the following environment variables:

* VESMGR_VNFNAME - VNF name as a string. Default: Vespa.
* VESMGR_NFNAMINGCODE - NF naming code as a string. Default: ricp.
* VESMGR_HB_INTERVAL - VES heartbeat interval as a string. For example: 30s.
* VESMGR_MEAS_INTERVAL - Measurement interval as a string. For example: 60s.
* VESMGR_PROMETHEUS_ADDR - Prometheus address. For example: http://127.0.0.1:123
* VESMGR_ALERTMANAGER_BIND_ADDR - Bind address to receive alerts from Prometheus AlertManager

* VESMGR_PRICOLLECTOR_ADDR - Primary collector FQDN as a string. For example: ricaux-entry.
* VESMGR_PRICOLLECTOR_PORT - Primary collector port id as an integer. Default: 8443.
* VESMGR_PRICOLLECTOR_SERVERROOT - Path before the /eventListener part of the POST URL as a string.
* VESMGR_PRICOLLECTOR_TOPIC - Primary collector topic as a string.
* VESMGR_PRICOLLECTOR_SECURE - Use HTTPS for VES collector. Possible string values: true or false.
* VESMGR_PRICOLLECTOR_USER - User name as a string.
* VESMGR_PRICOLLECTOR_PASSWORD - Password as a string.
* VESMGR_PRICOLLECTOR_PASSPHASE - Passphrase as a string.

* VESMGR_APPMGRDOMAIN - Application manager domain. This is for testing purposes, only. Default: service-ricplt-appmgr-http.ricplt.svc.cluster.local.

Liveness probe
==============

The VESPA manager replies to liveness HTTP GET at path /supervision.

Errors
======

The VESPA manager exits in the following error cases:

* The VES Agent exits
* An unrecoverable system error during the initialization, for example
  * Creation of the VES Agent configuration file fails
  * Creation of a HTTP request message fails

Unit Tests
==========

In order to run the VESPA manager unit tests, give the following command:

```shell
go test ./... -v
```

