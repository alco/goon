package main

import (
	"os"
)

func die(reason string) {
	logger.Println("dying:", reason)
	println(reason)
	os.Exit(-1)
}

func die_usage(reason string) {
	logger.Println("dying:", reason)
	println(reason)
	println(usage)
	os.Exit(-1)
}

func fatal(any interface{}) {
	logger.Panicln(any)
}

func fatal_if(any interface{}) {
	if any != nil {
		logger.Panicln(any)
	}
}

func shplit(str string) []string {
	// FIXME
	return []string{str}
}
