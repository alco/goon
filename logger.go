package main

import (
	"io"
	"log"
	"os"
)

var logger *log.Logger

type dummyWriter bool

func (w dummyWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func initLogger(flag string) {
	const kFileMode = 0666

	var file io.Writer
	switch flag {
	case "":
		file = new(dummyWriter)
	case "|1":
		file = os.Stdout
	case "|2":
		file = os.Stderr
	default:
		var err error
		file, err = os.OpenFile(flag, os.O_CREATE|os.O_WRONLY, 0666)
		fatal_if(err)
	}
	logger = log.New(file, "[goon]: ", log.Lmicroseconds)
}

