package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

var logFatalf = log.Fatalf

type messageEvent struct {
	Event   string `json:"event"`
	Content string `json:"content"`
}

type activityEvent struct {
	Event string `json:"event"`
	Title string `json:"title"`
}

type flowdockResponse struct {
	ThreadID string `json:"thread_id"`
}

func main() {
	apiURL := "https://api.flowdock.com/messages?flow_token="

	flowToken := os.Getenv("PLUGIN_FLOW_TOKEN")
	if flowToken == "" {
		log.Fatalln("Missing setting: flow_token")
	}
	flowURL := apiURL + flowToken

	message := os.Getenv("PLUGIN_MESSAGE")
	if message == "" {
		repoName := os.Getenv("DRONE_REPO")
		buildLink := os.Getenv("DRONE_BUILD_LINK")
		buildStatus := os.Getenv("DRONE_BUILD_STATUS")
		message = fmt.Sprintf("Status of build [%s](%s) is %s", repoName, buildLink, buildStatus)
	}

	eventType := os.Getenv("PLUGIN_MESSAGE_TYPE")

	if eventType == "activity" {
		msg := activityEvent{
			Event: "activity",
			Title: message,
		}

		raw, err := json.Marshal(msg)
		if err != nil {
			log.Fatalln(err)
		}

		postMessage(raw, flowURL)
	} else {
		msg := messageEvent{
			Event:   "message",
			Content: message,
		}

		raw, err := json.Marshal(msg)
		if err != nil {
			log.Fatalln(err)
		}

		postMessage(raw, flowURL)
	}
}

func postMessage(raw []byte, flowURL string) {
	req, err := http.NewRequest("POST", flowURL, bytes.NewReader(raw))
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-flowdock-wait-for-message", "true")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	if resp != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		messageThread := getThread(body)
		if resp.StatusCode == 201 {
			log.Println("Success! flowdock posted message to thread: " + messageThread)
		} else if resp.StatusCode == 202 {
			log.Println("Warning, flowdock didn't return thread id: " + resp.Status)
		} else {
			logFatalf("Failed to post message, flowdock api returned: %s", resp.Status)
		}
		resp.Body.Close()

		upload(client, flowURL, "coffee.gif", messageThread)
	}

	if err != nil {
		log.Fatalln(err)
	}
}

func upload(client *http.Client, url string, file string, thread string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	fileUpload := mustOpen(file)

	values := map[string]io.Reader{
		"content":   fileUpload,
		"thread_id": strings.NewReader(thread),
		"event":     strings.NewReader("file"),
	}

	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if x, ok := r.(*os.File); ok {
			fw, _ = w.CreateFormFile(key, x.Name())
		} else {
			fw, _ = w.CreateFormField(key)
		}
		io.Copy(fw, r)
	}
	w.Close()

	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	if res.StatusCode != http.StatusAccepted {
		logFatalf("Failed to post file: %s", res.Status)
	} else {
		log.Printf("Added file %s to thread: %s", fileUpload.Name(), thread)
	}
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

func getThread(body []byte) string {
	var s flowdockResponse
	err := json.Unmarshal(body, &s)
	if err != nil {
		log.Println("whoops:", err)
	}
	return s.ThreadID
}
