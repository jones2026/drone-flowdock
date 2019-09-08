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
	"path/filepath"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

var logFatalf = log.Fatalf
var openFile = os.Open
var settings pluginSettings
var apiURL = "https://api.flowdock.com/messages?flow_token="

type pluginSettings struct {
	Message   string `required:"true"`
	Files     string
	MaxFiles  int    `default:"5" split_words:"true"`
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

	msg := messageEvent{
		Event:   "message",
		Content: settings.Message,
	}

	raw, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln(err)
	}

	client := &http.Client{}
	messageThread := postMessage(client, raw)

	if settings.Files != "" {
		var filesUploaded = 0
		matches, _ := filepath.Glob(settings.Files)
		for _, filename := range matches {
			if filesUploaded < settings.MaxFiles {
				filesUploaded++
				fileUpload := mustOpen(filename)
				uploadFile(client, fileUpload, messageThread)
			} else {
				log.Printf("Maximum files (%d) to attach exceeded. Skipping remaining files.", settings.MaxFiles)
				break
			}
		}
	}

}

func fetchSettings() error {
	err := envconfig.Process("PLUGIN", &settings)
	return err
}

func getFlowdockRequest(b *bytes.Buffer) *http.Request {
	flowURL := apiURL + settings.FlowToken
	req, err := http.NewRequest("POST", flowURL, b)
	if err != nil {
		logFatalf(err.Error())
	}
	req.Header.Set("X-flowdock-wait-for-message", "true")
	return req
}

func postMessage(client *http.Client, raw []byte) string {
	b := bytes.NewBuffer(raw)
	req := getFlowdockRequest(b)
	req.Header.Set("Content-Type", "application/json")

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

func uploadFile(client *http.Client, file *os.File, thread string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	values := map[string]io.Reader{
		"content":   file,
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

	req := getFlowdockRequest(&b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	if res.StatusCode == http.StatusCreated {
		log.Printf("Added file %s to thread: %s", file.Name(), thread)
	} else {
		logFatalf("Failed to post file: %s", res.Status)
	}
}

func mustOpen(f string) *os.File {
	r, err := openFile(f)
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
