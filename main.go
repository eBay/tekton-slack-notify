// Copyright 2022 eBay Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/pflag"
)

var (
	tokenFile string

	channel  string
	threadTS string
	text     string
	tsFile   string
)

func init() {
	pflag.StringVar(&tokenFile, "token-file", "", "file path containing slack API token")
	pflag.StringVar(&channel, "channel", "", "channel ID to send message in")
	pflag.StringVar(&threadTS, "thread-ts", "", "thread ts, when specified, the message will be sent to thread")
	pflag.StringVar(&text, "text", "", "message text")
	pflag.StringVar(&tsFile, "ts-file", "", "if specified, the message ts will be written into this file")
}

// For details about API, please take a look at https://api.slack.com/methods/chat.postMessage

type message struct {
	Channel   string `json:"channel"`
	Text      string `json:"text"`
	ThreadTS  string `json:"thread_ts,omitempty"`
	LinkNames bool   `json:"link_names"`
}

type response struct {
	OK      bool   `json:"ok"`
	Channel string `json:"channel"`
	TS      string `json:"ts"`
}

func main() {
	pflag.Parse()

	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		log.Fatalf("failed to read %s: %s", tokenFile, err)
	}

	payload, err := json.Marshal(&message{
		Channel:   channel,
		Text:      text,
		ThreadTS:  threadTS,
		LinkNames: true,
	})
	if err != nil {
		log.Fatalf("failed to marshal message: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://slack.com/api/chat.postMessage", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalf("failed to new slack chat.postMessage request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(string(token)))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to do slack chat.postMessage request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Fatalf("failed to send slack message, unexpected status code: %d", resp.StatusCode)
	} else {
		log.Printf("message successfully sent")
	}

	if tsFile == "" {
		return
	}

	// extract the ts info from response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %s", err)
	}
	response := &response{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Fatalf("failed to unmarshal response: %s", err)
	}

	err = ioutil.WriteFile(tsFile, []byte(response.TS), 0644)
	if err != nil {
		log.Fatalf("failed to write (%s) into %s: %s", response.TS, tsFile, err)
	}
}
