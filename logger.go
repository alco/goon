package main

import (
	"log"
	"os"
)

var logger *log.Logger

func init() {
	var filename string
	if logsEnabled {
		filename = "goon.log"
	} else {
		filename = os.DevNull
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	logger = log.New(file, "goon", log.Lmicroseconds)
}

