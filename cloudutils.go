package common

import (
	"bytes"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"io"
	"log"
	"net/http"
	"os"
)

//connect without VPC connector https://cloud.google.com/sql/docs/mysql/connect-functions#go

const (
	EnvUrlEmailQueue = "URL_EMAIL_QUEUE"
	//test remove

	CloudfunctionSourceDir = "serverless_function_source_code"
)

func FixTemplateSourceDir() {
	fileInfo, err := os.Stat(CloudfunctionSourceDir)
	if err == nil && fileInfo.IsDir() {
		_ = os.Chdir(CloudfunctionSourceDir)
	}
}

func PostJson(alerts []Alert) {
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

func GetSecretsClient() *secretmanager.Client {
	// Create the client.
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to setup client: %v", err)
	}

	return client
}

func GetSecret(envSecretName string) string {
	if os.Getenv("LOCAL") == "1" {
		return getSecretLocal(envSecretName)
	} else {
		return getSecret(envSecretName)
	}
}

func getSecretLocal(envSecretName string) string {
	return os.Getenv(envSecretName)
}

func getSecret(envSecretName string) string {
	secretName := os.Getenv(envSecretName)
	client := GetSecretsClient()

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: "projects/investor-tracker/secrets/" + secretName + "/versions/latest",
	}

	ctx := context.Background()
	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		log.Fatalf("failed to access secret version: %v", err)
	}
	return string(result.Payload.Data)
}
