package util

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func Output(msg string) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("%s %s\n", bold("statsquid"), msg)
}

func OutputErr(e error) {
	bold := color.New(color.Bold).SprintFunc()
	fmt.Printf("%s %s\n", bold("statsquid"), e)
}

func FailOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
