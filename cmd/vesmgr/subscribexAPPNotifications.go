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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// appmgr API
const appmgrSubsPath = "/ric/v1/subscriptions"

var appmgrUrl = "http://" + appmgrDomain + ":" + appmgrPort + appmgrSubsPath

var errPostingFailed error = errors.New("Posting subscriptions failed")
var errWrongStatusCode error = errors.New("Wrong subscriptions response StatusCode")

func subscribexAppNotifications(targetUrl string, subscriptions chan subsChannel, timeout time.Duration) {
	requestBody := []byte(fmt.Sprintf(`{"maxRetries": 5, "retryTimer": 5, "eventType":"all", "targetUrl": "%v"}`, targetUrl))
	req, err := http.NewRequest("POST", appmgrUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Error("Setting NewRequest failed: %s", err)
		subscriptions <- subsChannel{false, err}
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	client.Timeout = time.Second * timeout
	for {
		err := subscribexAppNotificationsClientDo(req, client)
		if err == nil {
			break
		} else if err != errPostingFailed && err != errWrongStatusCode {
			subscriptions <- subsChannel{false, err}
			return
		}
		time.Sleep(3 * time.Second)
	}
	subscriptions <- subsChannel{true, nil}
}

func subscribexAppNotificationsClientDo(req *http.Request, client *http.Client) error {
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Posting subscriptions failed: %s", err)
		return errPostingFailed
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusCreated {
			logger.Info("Subscriptions response StatusCode: %d", resp.StatusCode)
			logger.Info("Subscriptions response headers: %s", resp.Header)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logger.Error("Subscriptions response Body read failed: %s", err)
				return err
			}
			logger.Info("Response Body: %s", body)
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(body), &result); err != nil {
				logger.Error("json.Unmarshal failed: %s", err)
				return err
			}
			logger.Info("Subscription id from the response: %s", result["id"].(string))
			vesmgr.appmgrSubsId = result["id"].(string)
			return nil
		} else {
			logger.Error("Wrong subscriptions response StatusCode: %d", resp.StatusCode)
			return errWrongStatusCode
		}
	}
}
