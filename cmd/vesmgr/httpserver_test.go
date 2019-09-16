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
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
)

type HttpServerTestSuite struct {
	suite.Suite
	listener       net.Listener
	ch_notif       chan []byte
	ch_supervision chan chan string
}

// suite setup creates the HTTP server
func (suite *HttpServerTestSuite) SetupSuite() {
	os.Unsetenv("http_proxy")
	os.Unsetenv("HTTP_PROXY")
	var err error
	suite.listener, err = net.Listen("tcp", ":0")
	suite.Nil(err)
	suite.ch_notif = make(chan []byte)
	suite.ch_supervision = make(chan chan string)
	startHttpServer(suite.listener, "/vesmgr_notif/", suite.ch_notif, suite.ch_supervision)
}

func (suite *HttpServerTestSuite) TestHtppServerSupervisionInvalidOperation() {
	resp, reply := suite.doPost("http://"+suite.listener.Addr().String()+SupervisionUrl, "supervision")
	suite.Equal("405 method not allowed\n", reply)
	suite.Equal(405, resp.StatusCode)
	suite.Equal("405 Method Not Allowed", resp.Status)
}

func (suite *HttpServerTestSuite) doGet(url string) (*http.Response, string) {
	resp, err := http.Get(url)
	suite.Nil(err)

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	return resp, string(contents)
}

func (suite *HttpServerTestSuite) doPost(serverUrl string, msg string) (*http.Response, string) {
	resp, err := http.Post(serverUrl, "data", strings.NewReader(msg))
	suite.Nil(err)

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	suite.Nil(err)
	return resp, string(contents)
}

func replySupervision(ch_supervision chan chan string, reply string) {
	ch_supervision_ack := <-ch_supervision
	ch_supervision_ack <- reply
}

func (suite *HttpServerTestSuite) TestHttpServerSupervision() {

	// start the "main loop" to reply to the supervision to the HTTPServer
	go replySupervision(suite.ch_supervision, "I'm just fine")

	resp, reply := suite.doGet("http://" + suite.listener.Addr().String() + SupervisionUrl)

	suite.Equal("I'm just fine", reply)
	suite.Equal(200, resp.StatusCode)
	suite.Equal("200 OK", resp.Status)
}

func (suite *HttpServerTestSuite) TestHttpServerInvalidUrl() {
	resp, reply := suite.doPost("http://"+suite.listener.Addr().String()+"/invalid_url", "foo")
	suite.Equal("404 page not found\n", reply)
	suite.Equal(404, resp.StatusCode)
	suite.Equal("404 Not Found", resp.Status)
}

func readXAppNotification(ch_notif chan []byte, ch chan []byte) {
	notification := <-ch_notif
	ch <- notification
}

func (suite *HttpServerTestSuite) TestHttpServerXappNotif() {
	// start the "main loop" to receive the xAppNotification message from the HTTPServer
	ch := make(chan []byte)
	go readXAppNotification(suite.ch_notif, ch)

	resp, reply := suite.doPost("http://"+suite.listener.Addr().String()+"/vesmgr_notif/", "test data")
	suite.Equal("", reply)
	suite.Equal(200, resp.StatusCode)
	suite.Equal("200 OK", resp.Status)
	notification := <-ch
	suite.Equal([]byte("test data"), notification)
}

func (suite *HttpServerTestSuite) TestHttpServerXappNotifInvalidOperation() {
	resp, reply := suite.doGet("http://" + suite.listener.Addr().String() + "/vesmgr_notif/")
	suite.Equal("405 method not allowed\n", reply)
	suite.Equal(405, resp.StatusCode)
	suite.Equal("405 Method Not Allowed", resp.Status)
}

func TestHttpServerSuite(t *testing.T) {
	suite.Run(t, new(HttpServerTestSuite))
}
