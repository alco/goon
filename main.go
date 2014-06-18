package main

import (
	"flag"
	"fmt"
	"os"
)

type protoImplT func(bool, bool, string, string, []string) error

const usage = "Usage: goon -proto <version> [options] -- <program> [<arg>...]"

var protoFlag = flag.String("proto", "", "protocol version (one of: 0.0)")
var inFlag  = flag.Bool("in", false, "enable reading from stdin")
var outFlag = flag.Bool("out", false, "output program's stdout")
var errFlag = flag.String("err", "nil", "output or redirect stderr")
var dirFlag = flag.String("dir", ".", "working directory for the spawned process")
var logFlag = flag.String("log", "", "enable logging")

func main() {
	flag.Parse()
	args := flag.Args()

	initLogger(*logFlag)
	validateOptsAndArgs(*protoFlag, args)

	/* Run external program and block until it terminates */
	err := findProtocolImpl(*protoFlag)(*inFlag, *outFlag, *errFlag, *dirFlag, args)

	/* Determine the exit status */
	if err != nil {
		//fmt.Printf("%#v\n", err)
		os.Exit(getExitStatus(err))
	}
}

func validateOptsAndArgs(protoFlag string, args []string) {
	if protoFlag == "" {
		die_usage("Please specify the protocol version.")
	}

	if len(args) < 1 {
		die_usage("Not enough arguments.")
	}

	logger.Printf("Flag values:\n  proto: %v\n  in: %v\n  out: %v\n  err: %v\n  dir: %v\nArgs: %v\n",
				  protoFlag, *inFlag, *outFlag, *errFlag, *dirFlag, args)
}

func findProtocolImpl(flag string) (impl protoImplT) {
	switch flag {
	case "0.0":
		impl = proto_0_0
	default:
		reason := fmt.Sprintf("Unsupported protocol version: %v", flag)
		die(reason)
	}
	return
}
