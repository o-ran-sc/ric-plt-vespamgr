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
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	mdcloggo "gerrit.o-ran-sc.org/r/com/golog.git"
)

var appmgrDomain string

const appmgrXAppConfigPath = "/ric/v1/config"
const appmgrPort = "8080"

type VesAgent struct {
	Pid     int
	name    string
	process *os.Process
}

type VesMgr struct {
	myIPAddress  string
	appmgrSubsId string
}

type subsChannel struct {
	subscribed bool
	err        error
}

var vesagent VesAgent
var vesmgr VesMgr
var logger *mdcloggo.MdcLogger

const vesmgrXappNotifPort = "8080"
const vesmgrXappNotifPath = "/vesmgr_xappnotif/"
const timeoutPostXAppSubscriptions = 5

func init() {
	logger, _ = mdcloggo.InitLogger("vesmgr")
}

func getMyIP() (myIP string, retErr error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logger.Error("net.InterfaceAddrs failed: %s", err.Error())
		return "", err
	}
	for _, addr := range addrs {
		// check the address type and if it is not a loopback take it
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				logger.Info("My IP Address: %s", ipnet.IP.String())
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", nil
}

func createConf(xappMetrics []byte) {
	f, err := os.Create("/etc/ves-agent/ves-agent.yaml")
	if err != nil {
		logger.Error("Cannot create vespa conf file: %s", err.Error())
		os.Exit(1)
	}
	defer f.Close()

	createVespaConfig(f, xappMetrics)
	logger.Info("Vespa config created")
}

func subscribeXAppNotifications(chSubscriptions chan subsChannel) {
	xappNotifUrl := "http://" + vesmgr.myIPAddress + ":" + vesmgrXappNotifPort + vesmgrXappNotifPath
	subsUrl := "http://" + appmgrDomain + ":" + appmgrPort + appmgrSubsPath
	go subscribexAppNotifications(xappNotifUrl, chSubscriptions, timeoutPostXAppSubscriptions, subsUrl)
	logger.Info("xApp notifications subscribed from %s", subsUrl)
}

func vesmgrInit() {
	vesagent.name = "ves-agent"
	logger.Info("vesmgrInit")

	var err error
	if vesmgr.myIPAddress, err = getMyIP(); err != nil || vesmgr.myIPAddress == "" {
		logger.Error("Cannot get myIPAddress: IP %s, err %s", vesmgr.myIPAddress, err.Error())
		return
	}

	var ok bool
	appmgrDomain, ok = os.LookupEnv("VESMGR_APPMGRDOMAIN")
	if ok {
		logger.Info("Using appmgrdomain %s", appmgrDomain)
	} else {
		appmgrDomain = "service-ricplt-appmgr-http.ricplt.svc.cluster.local"
		logger.Info("Using default appmgrdomain %s", appmgrDomain)
	}
	chXAppSubscriptions := make(chan subsChannel)
	chXAppNotifications := make(chan []byte)
	chSupervision := make(chan chan string)
	chVesagent := make(chan error)

	listener, err := net.Listen("tcp", vesmgr.myIPAddress+":"+vesmgrXappNotifPort)
	startHttpServer(listener, vesmgrXappNotifPath, chXAppNotifications, chSupervision)

	subscribeXAppNotifications(chXAppSubscriptions)

	createConf([]byte{})
	startVesagent(chVesagent)

	runVesmgr(chVesagent, chSupervision, chXAppNotifications, chXAppSubscriptions)
}

func startVesagent(ch chan error) {
	cmd := exec.Command(vesagent.name, "-i", os.Getenv("VESMGR_HB_INTERVAL"), "-m", os.Getenv("VESMGR_MEAS_INTERVAL"), "--Measurement.Prometheus.Address", os.Getenv("VESMGR_PROMETHEUS_ADDR"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		logger.Error("vesmgr exiting, ves-agent start failed: %s", err)
		go func() {
			ch <- err
		}()
	} else {
		logger.Info("ves-agent started with pid %d", cmd.Process.Pid)
		vesagent.Pid = cmd.Process.Pid
		vesagent.process = cmd.Process
		go func() {
			// wait ves-agent exit and then post the error to the channel
			err := cmd.Wait()
			ch <- err
		}()
	}
}

func killVespa(process *os.Process) {
	logger.Info("Killing vespa")
	err := process.Kill()
	if err != nil {
		logger.Error("Cannot kill vespa: %s", err.Error())
	}
}

func queryXAppsStatus(appmgrUrl string, timeout time.Duration) ([]byte, error) {

	logger.Info("query xAppStatus started, url %s", appmgrUrl)
	req, err := http.NewRequest("GET", appmgrUrl, nil)
	if err != nil {
		logger.Error("Failed to create a HTTP request: %s", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	client.Timeout = time.Second * timeout
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Query xApp status failed: %s", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Failed to read xApp status body: %s", err)
			return nil, err
		}
		logger.Info("query xAppStatus completed")
		return body, nil
	} else {
		logger.Error("Error from xApp status query: %s", resp.Status)
		return nil, errors.New(resp.Status)
	}
}

type state int

const (
	normalState           state = iota
	vespaTerminatingState state = iota
)

func queryConf() ([]byte, error) {
	return queryXAppsStatus("http://"+appmgrDomain+":"+appmgrPort+appmgrXAppConfigPath,
		10*time.Second)
}

func runVesmgr(chVesagent chan error, chSupervision chan chan string, chXAppNotifications chan []byte, chXAppSubscriptions chan subsChannel) {

	logger.Info("vesmgr main loop ready")
	mystate := normalState
	var xappStatus []byte
	for {
		select {
		case supervision := <-chSupervision:
			logger.Info("vesmgr: supervision")
			supervision <- "OK"
		case xAppNotif := <-chXAppNotifications:
			logger.Info("vesmgr: xApp notification")
			logger.Info(string(xAppNotif))
			/*
			 * If xapp status query fails then we cannot create
			 * a new configuration and kill vespa.
			 * In that case we assume that
			 * the situation is fixed when the next
			 * xapp notif comes
			 */
			var err error
			xappStatus, err = queryConf()
			if err == nil {
				killVespa(vesagent.process)
				mystate = vespaTerminatingState
			}
		case err := <-chVesagent:
			switch mystate {
			case vespaTerminatingState:
				logger.Info("Vesagent termination completed")
				createConf(xappStatus)
				startVesagent(chVesagent)
				mystate = normalState
			default:
				logger.Error("Vesagent exited: " + err.Error())
				os.Exit(1)
			}
		case isSubscribed := <-chXAppSubscriptions:
			if isSubscribed.err != nil {
				logger.Error("Failed to make xApp subscriptions, vesmgr exiting: %s", isSubscribed.err)
				os.Exit(1)
			}
		}
	}
}
