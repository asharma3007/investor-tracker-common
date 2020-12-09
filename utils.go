package common

import (
	"fmt"
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
