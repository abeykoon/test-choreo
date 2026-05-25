package main

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "", 0)

var connections = []string{
	"CHOREO_TESTER_TO_ORG",
	"CHOREO_TESTER_TO_PUBLIC",
	"CHOREO_TESTER_TO_PROJECT",
}

var suffixes = []string{
	"SERVICEURL",
	"CHOREOAPIKEY",
	"CONSUMERKEY",
	"CONSUMERSECRET",
	"TOKENURL",
}

func main() {
	for _, prefix := range connections {
		logger.Printf("[%s]", prefix)
		for _, suffix := range suffixes {
			key := prefix + "_" + suffix
			val := os.Getenv(key)
			if val == "" {
				logger.Printf("  %s = (not set)", key)
			} else {
				logger.Printf("  %s = %s", key, val)
			}
		}
	}
}
