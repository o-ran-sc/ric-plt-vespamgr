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
	"os/exec"
	mdcloggo "gerrit.o-ran-sc.org/r/com/golog.git"
)

type VesAgent struct {
	Pid  int
	name string
}

var vesagent VesAgent
var logger *mdcloggo.MdcLogger
var osExit = os.Exit

func init() {
	logger, _ = mdcloggo.InitLogger("vesmgr")
}

/* Function to initialize vesmgr */
func vesmgrInit() {
	vesagent.name = "ves-agent"
	logger.MdcAdd("vesmgr", "0.0.1")
	logger.Info("vesmgrInit")

	/* Subscribe notifications from xAPP Mgr */
	//subscribexAppNotifications()

	// create configuration
	f, err := os.Create("/etc/ves-agent/ves-agent.yaml")
	if err != nil {
		logger.Error("Cannot create vespa conf file: %s", err.Error())
		return
	}
	defer f.Close()

	createVespaConfig(f)

	/* Start ves-agent */
	ch := startVesagent()

	runVesmgr(ch)
}

func startVesagent() chan error {
	/* Start ves-agent */
	cmd := exec.Command(vesagent.name, "-i", os.Getenv("VESMGR_HB_INTERVAL"), "-m", os.Getenv("VESMGR_MEAS_INTERVAL"), "-f", os.Getenv("VESMGR_PRICOLLECTOR_ADDR"), "-p", os.Getenv("VESMGR_PRICOLLECTOR_PORT"), "--Measurement.Prometheus.Address", os.Getenv("VESMGR_PROMETHEUS_ADDR"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	ch := make(chan error)
	if err := cmd.Start(); err != nil {
		logger.Error("vesmgr exiting, ves-agent start failed: %s", err)
		go func() {
			ch <- err
		}()
	} else {
		logger.Info("ves-agent started with pid %d", cmd.Process.Pid)
		vesagent.Pid = cmd.Process.Pid
		go func() {
			// wait ves-agent exit and then post the error to the channel
			err := cmd.Wait()
			ch <- err
		}()
	}

	return ch
}

func runVesmgr(ch chan error) {
	for {
		err := <-ch
		logger.Error("Vesagent exited: " + err.Error())
		osExit(1)
	}
}
