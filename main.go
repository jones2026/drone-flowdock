package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func postMessage(msg, flowURL string) {
	raw, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := http.Post(flowURL, "application/json", bytes.NewReader(raw))
	if resp != nil {
		fmt.Println(resp.Status)
		resp.Body.Close()
	}

	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	apiURL := "https://api.flowdock.com/messages?flow_token="

	flowToken := os.Getenv("PLUGIN_FLOW_TOKEN")
	if flowToken == "" {
		log.Fatalln("Missing flow token")
	}
	flowURL := apiURL + flowToken

	message := os.Getenv("PLUGIN_MESSAGE")
	if message == "" {
		repoName := os.Getenv("DRONE_REPO")
		buildLink := os.Getenv("DRONE_BUILD_LINK")
		buildStatus := os.Getenv("DRONE_BUILD_STATUS")
		message = fmt.Sprintf("Status of build [%s](%s) is %s", repoName, buildLink, buildStatus)
	}

	messageType := os.Getenv("PLUGIN_MESSAGE_TYPE")

	if messageType == "activity" {
		msg := struct {
			Event   string `json:"event"`
			Title string `json:"title"`
		}{
			Event:   "activity",
			Title: message,
		}

		postMessage(msg, flowURL)
	} else {
		msg := struct {
			Event   string `json:"event"`
			Content string `json:"content"`
		}{
			Event:   "message",
			Content: message,
		}

		postMessage(msg, flowURL)
	}
}
