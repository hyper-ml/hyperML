package main

import (
	"fmt"
	"os"
	"strings"
)

func exitWithError(format string, args ...interface{}) {
	if errString := strings.TrimSpace(fmt.Sprintf(format, args...)); errString != "" {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", errString)
	}
	os.Exit(1)
}

func exitWithSuccess() {
	os.Exit(0)
}

func main() {
	var err error

	rootCommand, err := RootCmd()
	if err != nil {
		exitWithError(err.Error())
	}

	rootCommand.Execute()

}
