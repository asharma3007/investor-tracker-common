package common

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func LogDebug(message string) {
	if os.Getenv("DEBUG") == "1" {
		fmt.Println(message)
	}
}

func Log(message string) {
	fmt.Println(message)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func Date(day int, month int, year int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}


func WriteStringToFile(filename, str string) {
	_ : os.Remove(filename)

	file, _ := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	datawriter := bufio.NewWriter(file)

	_, _ = datawriter.WriteString(str)

	_ : datawriter.Flush()
	_ : file.Close()
}

func PostJsonToUrl(url string, object interface{}) {
	buf := new(bytes.Buffer)
	_ = json.NewEncoder(buf).Encode(object)
	req, _ := http.NewRequest("POST", url, buf)

	client := &http.Client{}
	res, err := client.Do(req)
	CheckError(err)

	defer res.Body.Close()

	fmt.Println("response Status:", res.Status)
	// Print the body to the stdout
	_, _ = io.Copy(os.Stdout, res.Body)
}
