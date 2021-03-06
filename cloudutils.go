package common

import (
	"bytes"
	"cloud.google.com/go/pubsub"
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
	Log("Getting secret " + envSecretName)
	if len(envSecretName) == 0 { return "" }

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
	if len(secretName) == 0 { return ""}

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

func PubSubPublish(projectID, topicID, msg string) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}

	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})
	// Block until the result is returned and a server-generated
	// ID is returned for the published message.
	id, err := result.Get(ctx)
	if err != nil {
		return fmt.Errorf("Get: %v", err)
	}
	Log(fmt.Sprintf("Published a message; msg ID: %v", id))
	return nil
}

func SendEmail(email MessageSendEmail) {
	jsonBytes, err := json.Marshal(email)
	CheckError(err)
	jsonStr := string(jsonBytes)

	Log(fmt.Sprintf("JSON to publish: %v", jsonStr))
	if os.Getenv("LOCAL") == "1" {
		WriteStringToFile("output/email.json", jsonStr)
		url := os.Getenv(EnvUrlEmailQueue)
		PostJsonToUrl(url, email)
	} else {
		err := PubSubPublish("investor-tracker", "alerts-ready", jsonStr)
		CheckError(err)
	}
}
