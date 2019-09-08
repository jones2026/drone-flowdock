package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

var logFatalf = log.Fatalf
var settings pluginSettings

type pluginSettings struct {
	Message   string `required:"true"`
	File      string
	FlowToken string `required:"true" split_words:"true"`
}

type messageEvent struct {
	Event   string `json:"event"`
	Content string `json:"content"`
}

type flowdockResponse struct {
	ThreadID string `json:"thread_id"`
}

func main() {

	err := fetchSettings()
	if err != nil {
		logFatalf(err.Error())
	}

	apiURL := "https://api.flowdock.com/messages?flow_token="
	flowURL := apiURL + settings.FlowToken

	msg := messageEvent{
		Event:   "message",
		Content: settings.Message,
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln(err)
	}

	client := &http.Client{}
	messageThread := postMessage(client, raw, flowURL)

	if settings.File != "" {
		upload(client, flowURL, settings.File, messageThread)
	}

}

func fetchSettings() error {
	err := envconfig.Process("PLUGIN", &settings)
	return err
}

func postMessage(client *http.Client, raw []byte, flowURL string) string {
	req, _ := http.NewRequest("POST", flowURL, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-flowdock-wait-for-message", "true")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	var messageThread string
	if resp != nil {
		body, _ := ioutil.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusCreated {
			messageThread = getThread(body)
			log.Println("Flowdock message posted to thread: " + messageThread)
		} else {
			logFatalf("Failed to post message, flowdock api returned: %s", resp.Status)
		}
		resp.Body.Close()
	}

	if err != nil {
		log.Fatalln(err)
	}
	return messageThread
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
		if file, ok := r.(io.Closer); ok {
			defer file.Close()
		}
		if file, ok := r.(*os.File); ok {
			fw, _ = w.CreateFormFile(key, file.Name())
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
	req.Header.Set("X-flowdock-wait-for-message", "true")

	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	if res.StatusCode == http.StatusCreated {
		log.Printf("Added file %s to thread: %s", fileUpload.Name(), thread)
	} else {
		logFatalf("Failed to post file: %s", res.Status)
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
