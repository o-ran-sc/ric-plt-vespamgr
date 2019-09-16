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
	"io/ioutil"
	"net"
	"net/http"
)

const SupervisionUrl = "/supervision/"

func startHttpServer(listener net.Listener, xappnotifUrl string, notif_ch chan []byte, supervision_ch chan chan string) {
	go runHttpServer(listener, xappnotifUrl, notif_ch, supervision_ch)
}

func runHttpServer(listener net.Listener, xappNotifUrl string, notif_ch chan []byte, supervision_ch chan chan string) {

	logger.Info("vesmgr http server serving at %s", listener.Addr())

	http.HandleFunc(xappNotifUrl, func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "POST":
			logger.Info("httpServer: POST in %s", xappNotifUrl)
			body, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				logger.Error("httpServer: Invalid body in POST request")
				return
			}
			notif_ch <- body
			return
		default:
			logger.Error("httpServer: Invalid method %s to %s", r.Method, r.URL.Path)
			http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})

	http.HandleFunc(SupervisionUrl, func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case "GET":
			logger.Info("httpServer: GET supervision")
			supervision_ack_ch := make(chan string)
			// send supervision to the main loop
			supervision_ch <- supervision_ack_ch
			reply := <-supervision_ack_ch
			logger.Info("httpServer: supervision ack from the main loop: %s", reply)
			fmt.Fprintf(w, reply)
			return
		default:
			logger.Error("httpServer: invalid method %s to %s", r.Method, r.URL.Path)
			http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
			return
		}

	})

	http.Serve(listener, nil)
}
