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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type AppmgrHttpServerTestSuite struct {
	suite.Suite
	savedAppmgrUrl string
	subscriptions  chan subsChannel
	xappNotifUrl   string
}

// suite setup
func (suite *AppmgrHttpServerTestSuite) SetupSuite() {
	vesmgr.appmgrSubsId = string("")
	vesmgr.myIPAddress, _ = getMyIP()
	suite.savedAppmgrUrl = appmgrUrl
	suite.xappNotifUrl = "http://" + vesmgr.myIPAddress + ":" + vesmgrXappNotifPort + vesmgrXappNotifPath
	suite.subscriptions = make(chan subsChannel)
}

// test setup
func (suite *AppmgrHttpServerTestSuite) SetupTest() {
	suite.subscriptions = make(chan subsChannel)
}

// test teardown
func (suite *AppmgrHttpServerTestSuite) TearDownTest() {
	appmgrUrl = suite.savedAppmgrUrl
	vesmgr.appmgrSubsId = string("")
}

func (suite *AppmgrHttpServerTestSuite) TestSubscribexAppNotifications() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		body, _ := ioutil.ReadAll(req.Body)
		var result map[string]interface{}
		err := json.Unmarshal([]byte(body), &result)
		suite.Nil(err)
		suite.Equal(5, int(result["maxRetries"].(float64)))
		suite.Equal(5, int(result["retryTimer"].(float64)))
		suite.Equal("all", result["eventType"].(string))
		suite.Equal("POST", req.Method)
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(`{"id":"deadbeef1234567890", "version":0, "eventType":"all"}`))
	}))
	defer testServer.Close()

	appmgrUrl = testServer.URL

	go subscribexAppNotifications(suite.xappNotifUrl, suite.subscriptions, 1)
	isSubscribed := <-suite.subscriptions
	suite.Nil(isSubscribed.err)
	suite.Equal("deadbeef1234567890", vesmgr.appmgrSubsId)
}

func (suite *AppmgrHttpServerTestSuite) TestSubscribexAppNotificationsWrongStatus() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusUnauthorized)
		res.Write([]byte(`{"id":"deadbeef1234567890", "version":0, "eventType":"all"}`))
	}))
	defer testServer.Close()

	appmgrUrl = testServer.URL
	requestBody := []byte(fmt.Sprintf(`{"maxRetries": 5, "retryTimer": 5, "eventType":"all", "targetUrl": "%v"}`, suite.xappNotifUrl))
	req, _ := http.NewRequest("POST", appmgrUrl, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	err := subscribexAppNotificationsClientDo(req, client)
	suite.Equal(errWrongStatusCode, err)
	// after failed POST vesmgr.appmgrSubsId holds an initial values
	suite.Equal("", vesmgr.appmgrSubsId)
}

func (suite *AppmgrHttpServerTestSuite) TestSubscribexAppNotificationsWrongUrl() {
	// use fake appmgrUrl that is not served in unit test
	appmgrUrl = "/I_do_not_exist/"
	requestBody := []byte(fmt.Sprintf(`{"maxRetries": 5, "retryTimer": 5, "eventType":"all", "targetUrl": "%v"}`, suite.xappNotifUrl))
	req, _ := http.NewRequest("POST", appmgrUrl, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}

	err := subscribexAppNotificationsClientDo(req, client)
	suite.Equal(errPostingFailed, err)
	// after failed POST vesmgr.appmgrSubsId holds an initial values
	suite.Equal("", vesmgr.appmgrSubsId)
}

func (suite *AppmgrHttpServerTestSuite) TestSubscribexAppNotificationsReadBodyFails() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Length", "1")
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
	}))
	defer testServer.Close()

	appmgrUrl = testServer.URL
	go subscribexAppNotifications(suite.xappNotifUrl, suite.subscriptions, 1)
	isSubscribed := <-suite.subscriptions
	suite.Equal("unexpected EOF", isSubscribed.err.Error())
	suite.Equal("", vesmgr.appmgrSubsId)
}

func (suite *AppmgrHttpServerTestSuite) TestSubscribexAppNotificationsUnMarshalFails() {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(`{""dump for UT": make(chan int),"}`))
	}))
	defer testServer.Close()

	appmgrUrl = testServer.URL
	go subscribexAppNotifications(suite.xappNotifUrl, suite.subscriptions, 1)
	isSubscribed := <-suite.subscriptions
	suite.Equal("invalid character 'd' after object key", isSubscribed.err.Error())
	suite.Equal("", vesmgr.appmgrSubsId)
}

func TestAppmgrHttpServerTestSuite(t *testing.T) {
	suite.Run(t, new(AppmgrHttpServerTestSuite))
}
