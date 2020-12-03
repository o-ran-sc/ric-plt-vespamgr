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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	app "gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/stretchr/testify/suite"
)

type VespaMgrTestSuite struct {
	suite.Suite
	vespaMgr *VespaMgr
}

// suite setup
func (suite *VespaMgrTestSuite) SetupSuite() {
	os.Unsetenv("http_proxy")
	os.Unsetenv("HTTP_PROXY")
	suite.vespaMgr = NewVespaMgr()
}

func (suite *VespaMgrTestSuite) TestSubscribexAppNotifications() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		log.Println("TestSubscribexAppNotifications ...")
		body, _ := ioutil.ReadAll(req.Body)
		var result map[string]interface{}
		err := json.Unmarshal([]byte(body), &result)
		suite.Nil(err)
		data := result["Data"].(map[string]interface{})
		suite.Equal(5, int(data["maxRetries"].(float64)))
		suite.Equal(5, int(data["retryTimer"].(float64)))
		suite.Equal("all", data["eventType"].(string))
		suite.Equal("POST", req.Method)
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(`{"id":"deadbeef1234567890", "version":0, "eventType":"all"}`))
	}))
	defer testServer.Close()

	suite.vespaMgr.SubscribeXappNotif(testServer.URL)
	suite.Equal("deadbeef1234567890", suite.vespaMgr.subscriptionId)
}

func (suite *VespaMgrTestSuite) TestSubscribexAppNotificationsWrongStatus() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusUnauthorized)
		res.Write([]byte(`{"id":"deadbeef1234567890", "version":0, "eventType":"all"}`))
	}))
	defer testServer.Close()

	requestBody := []byte(fmt.Sprintf(`{"maxRetries": 5, "retryTimer": 5, "eventType":"all", "targetUrl": "%v"}`, "localhost:8080"))
	req, _ := http.NewRequest("POST", testServer.URL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	suite.vespaMgr.subscriptionId = ""
	suite.vespaMgr.DoSubscribe(testServer.URL, requestBody)
	suite.Equal("", suite.vespaMgr.subscriptionId)
}

func (suite *VespaMgrTestSuite) TestSubscribexAppNotificationsReadBodyFails() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Length", "1")
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
	}))
	defer testServer.Close()

	suite.vespaMgr.subscriptionId = ""
	suite.vespaMgr.DoSubscribe(testServer.URL, []byte{})
	suite.Equal("", suite.vespaMgr.subscriptionId)
}

func (suite *VespaMgrTestSuite) TestSubscribexAppNotificationsUnMarshalFails() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(`{""dump for UT": make(chan int),"}`))
	}))
	defer testServer.Close()

	suite.vespaMgr.subscriptionId = ""
	suite.vespaMgr.DoSubscribe(testServer.URL, []byte{})
	suite.Equal("", suite.vespaMgr.subscriptionId)
}

func (suite *VespaMgrTestSuite) TestQueryXAppsConfigOk() {
	listener, err := net.Listen("tcp", ":0")
	suite.Nil(err)

	http.HandleFunc("/test_url/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			fmt.Fprintf(w, "reply message")
		}
	})

	go http.Serve(listener, nil)

	xappConfig, err := suite.vespaMgr.QueryXappConf("http://" + listener.Addr().String() + "/test_url/")
	log.Println("####TestQueryXAppsConfigOk ...", xappConfig, err)

	suite.NotNil(xappConfig)
	suite.Nil(err)
	suite.Equal(xappConfig, []byte("reply message"))
}

func (suite *VespaMgrTestSuite) TestHandlexAppNotification() {
	data, err := ioutil.ReadFile("../../test/xApp_config_test_output.json")
	suite.Nil(err)

	pbodyEn, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", "/ric/v1/xappnotif", bytes.NewBuffer(pbodyEn))
	handleFunc := http.HandlerFunc(suite.vespaMgr.HandlexAppNotification)
	response := executeRequest(req, handleFunc)
	suite.Equal(http.StatusOK, response.Code)
}

func (suite *VespaMgrTestSuite) TestSubscribexAppNotificationsOnStartup() {
	suite.vespaMgr.Run(false, false)
	time.Sleep(2 * time.Second)

	suite.vespaMgr.Consume(&app.RMRParams{})
	suite.vespaMgr.StatusCB()
	suite.vespaMgr.rmrReady = true
	suite.vespaMgr.StatusCB()
}

func executeRequest(req *http.Request, handleR http.HandlerFunc) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handleR.ServeHTTP(rr, req)
	return rr
}

func TestVespaMgrTestSuite(t *testing.T) {
	suite.Run(t, new(VespaMgrTestSuite))
}
