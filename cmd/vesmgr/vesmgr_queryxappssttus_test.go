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
	"fmt"
	"github.com/stretchr/testify/suite"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

type do func(w http.ResponseWriter)

type QueryXAppsStatusTestSuite struct {
	suite.Suite
	listener    net.Listener
	xAppMgrFunc do
}

// suite setup creates the HTTP server
func (suite *QueryXAppsStatusTestSuite) SetupSuite() {
	os.Unsetenv("http_proxy")
	os.Unsetenv("HTTP_PROXY")
	var err error
	suite.listener, err = net.Listen("tcp", ":0")
	suite.Nil(err)
	go runXAppMgr(suite.listener, "/test_url/", suite)
}

func runXAppMgr(listener net.Listener, url string, suite *QueryXAppsStatusTestSuite) {

	http.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			suite.xAppMgrFunc(w)
		}
	})
	http.Serve(listener, nil)
}

func (suite *QueryXAppsStatusTestSuite) TestQueryXAppsStatusFailsWithTimeout() {
	do_sleep := func(w http.ResponseWriter) {
		time.Sleep(time.Second * 2)
	}
	suite.xAppMgrFunc = do_sleep

	data, err := queryXAppsStatus("http://"+suite.listener.Addr().String()+"/test_url/", 1)
	suite.Nil(data)
	suite.NotNil(err)
	e, ok := err.(*url.Error)
	suite.Equal(ok, true)
	suite.Equal(e.Timeout(), true)
}

func (suite *QueryXAppsStatusTestSuite) TestQueryXAppsStatusFailsWithAnErrorReply() {
	do_reply_with_err := func(w http.ResponseWriter) {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
	}
	suite.xAppMgrFunc = do_reply_with_err

	data, err := queryXAppsStatus("http://"+suite.listener.Addr().String()+"/test_url/", 1)
	suite.Nil(data)
	suite.NotNil(err)
	suite.Equal("405 Method Not Allowed", err.Error())
}

func (suite *QueryXAppsStatusTestSuite) TestQueryXAppsStatusOk() {
	do_reply := func(w http.ResponseWriter) {
		fmt.Fprintf(w, "reply message")
	}
	suite.xAppMgrFunc = do_reply

	data, err := queryXAppsStatus("http://"+suite.listener.Addr().String()+"/test_url/", 1)
	suite.NotNil(data)
	suite.Nil(err)
	suite.Equal(data, []byte("reply message"))
}

func TestQueryXAppsStatusTestSuite(t *testing.T) {
	suite.Run(t, new(QueryXAppsStatusTestSuite))
}
