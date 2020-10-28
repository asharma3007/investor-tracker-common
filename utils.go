package common

import "fmt"

func Log(message string) {
	fmt.Println(message)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}
