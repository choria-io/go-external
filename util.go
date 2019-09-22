package agent

import (
	"fmt"
	"os"
)

func fileExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

// Errorf produce an error level message
func Errorf(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
}

// Infof produce an info level message
func Infof(format string, a ...interface{}) {
	fmt.Printf(format+"\n", a...)
}
