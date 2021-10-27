package cli

import (
	"fmt"
)

func Successf(template string, args ...interface{}) {
	fmt.Printf("SUCCESS: " + template + "\n", args...)
}

func Warnf(template string, args ...interface{}) {
	fmt.Printf("WARNING: " + template + "\n", args...)
}

func Errorf(template string, args ...interface{}) {
	fmt.Printf("ERROR: " + template + "\n", args...)
}

func Wrapf(err error, template string, args ...interface{}) {
	fmt.Printf("FAILURE: %s: %+v\n", fmt.Sprintf(template, args...), err)
}
