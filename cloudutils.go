package investor_tracker_common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	EnvUrlEmailQueue 	= "URL_EMAIL_QUEUE"

	CloudfunctionSourceDir = "serverless_function_source_code"
)

func FixTemplateSourceDir() {
	fileInfo, err := os.Stat(CloudfunctionSourceDir)
	if err == nil && fileInfo.IsDir() {
		_ = os.Chdir(CloudfunctionSourceDir)
	}
}


func PostJson(alerts []alert) {
	url := os.Getenv(EnvUrlEmailQueue)

	buf := new(bytes.Buffer)
	_ = json.NewEncoder(buf).Encode(alerts)
	req, _ := http.NewRequest("POST", url, buf)

	client := &http.Client{}
	res, err := client.Do(req)
	CheckError(err)

	defer res.Body.Close()

	fmt.Println("response Status:", res.Status)
	// Print the body to the stdout
	_, _ = io.Copy(os.Stdout, res.Body)
}
