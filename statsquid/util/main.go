package util

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func Output(s string, a ...interface{}) {
	msg := fmt.Sprintf(s, a)
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("%s %s\n", bold("statsquid"), msg)
}

func OutputErr(s string, a ...interface{}) {
	msg := fmt.Errorf(s, a)
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("%s %s\n", bold("statsquid"), msg)
}

func FailOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
