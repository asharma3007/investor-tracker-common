package common

import (
	"fmt"
	"os"
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
