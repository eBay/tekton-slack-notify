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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	reaction string
	debug    bool
)

func init() {
	pflag.StringVar(&tokenFile, "token-file", "", "file path containing slack API token")
	pflag.StringVar(&channel, "channel", "", "channel ID to send message in")
	pflag.StringVar(&threadTS, "thread-ts", "", "thread ts, when specified, the message will be sent to thread")
	pflag.StringVar(&text, "text", "", "message text")
	pflag.StringVar(&tsFile, "ts-file", "", "if specified, the message ts will be written into this file")
	pflag.StringVar(&reaction, "reaction", "", "emoji reaction to add to the message")
	pflag.BoolVar(&debug, "debug", false, "pass to debug responses")
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

func sendSlackRequest(token, url string, payload []byte, debug bool) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed: %s", string(bodyBytes))
	}

	if debug {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Println(string(bodyBytes))
	}

	return resp, nil
}

func publishMessage(token string, msg *message, tsFile string, debug bool) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	resp, err := sendSlackRequest(token, "https://slack.com/api/chat.postMessage", payload, debug)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Printf("message successfully sent")

	if tsFile == "" {
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	response := &response{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	err = os.WriteFile(tsFile, []byte(response.TS), 0644)
	if err != nil {
		return fmt.Errorf("failed to write (%s) into %s: %w", response.TS, tsFile, err)
	}

	return nil
}

func addReaction(token, channel, threadTS, reaction string, debug bool) error {
	payload := map[string]string{
		"channel":   channel,
		"timestamp": threadTS,
		"name":      reaction,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal reaction payload: %w", err)
	}

	resp, err := sendSlackRequest(token, "https://slack.com/api/reactions.add", payloadBytes, debug)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func main() {
	pflag.Parse()

	token, err := os.ReadFile(tokenFile)
	if err != nil {
		log.Fatalf("failed to read %s: %s", tokenFile, err)
	}

	msg := &message{
		Channel:   channel,
		Text:      text,
		ThreadTS:  threadTS,
		LinkNames: true,
	}

	if text != "" {
		err = publishMessage(strings.TrimSpace(string(token)), msg, tsFile, debug)
		if err != nil {
			log.Fatalf("failed to publish message: %s", err)
		}
	}

	if reaction != "" && threadTS != ""{
		err = addReaction(strings.TrimSpace(string(token)), channel, threadTS, reaction, debug)
		if err != nil {
			log.Fatalf("failed to add reaction: %s", err)
		}
	}
}
