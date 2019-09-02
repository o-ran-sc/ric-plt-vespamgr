# RIC VESPA manager

The VESPA manager uses the VES Agent (https://github.com/nokia/ONAP-VESPA)
to adapt near-RT RIC internal statistics' collection using Prometheus 
(xApps and platform containers) to ONAP's VES (VNF event streaming).

The VESPA manager starts and configures the VES Agent.

# Environment variables

The VESPA manager container requires the following environment variables:

* VESMGR_HB_INTERVAL - VES heartbeat interval as a string. For example: 30s.
* VESMGR_MEAS_INTERVAL - Measurement interval as a string. For example: 60s.
* VESMGR_PROMETHEUS_ADDR - Prometheus address. For example: http://127.0.0.1:123

* VESMGR_PRICOLLECTOR_ADDR - Primary collector FQDN as a string. For example: ricaux-entry.
* VESMGR_PRICOLLECTOR_PORT - Primary collector port id as an integer. Default: 8443.
* VESMGR_PRICOLLECTOR_SERVERROOT - Path before the /eventListener part of the POST URL as a string.
* VESMGR_PRICOLLECTOR_TOPIC - Primary collector topic as a string.
* VESMGR_PRICOLLECTOR_SECURE - Use HTTPS for VES collector. Possible string values: true or false.
* VESMGR_PRICOLLECTOR_USER - User name as a string.
* VESMGR_PRICOLLECTOR_PASSWORD - Password as a string.
* VESMGR_PRICOLLECTOR_PASSPHASE - Passphrase as a string.

# Unit Tests

In order to run the VESPA manager unit tests, give the following command:

```
go test ./... -v
```

# License

See [LICENSES.txt](LICENSES.txt) file.
