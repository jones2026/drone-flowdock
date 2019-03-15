package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	apiURL := "https://api.flowdock.com/messages?flow_token="

	flowToken := os.Getenv("PLUGIN_FLOW_TOKEN")
	if flowToken == "" {
		log.Fatalln("Missing flow token")
	}

	message := os.Getenv("PLUGIN_MESSAGE")
	if message == "" {
		// repoName := os.Getenv("DRONE_REPO")
		// buildLink := os.Getenv("PLUGIN_FLOW_TOKEN")
		// buildStatus := os.Getenv("DRONE_STATUS")
		message = "Drone build passed"
	}

	msg := struct {
		Event   string `json:"event"`
		Content string `json:"content"`
	}{
		Event:   "message",
		Content: message,
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln(err)
	}

	flowURL := apiURL + flowToken
	resp, err := http.Post(flowURL, "application/json", bytes.NewReader(raw))
	if resp != nil {
		fmt.Println(resp.Status)
		resp.Body.Close()
	}

	if err != nil {
		log.Fatalln(err)
	}
}
