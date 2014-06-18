package main

import (
	"os"
	"os/exec"
)

func proto_0_0(inFlag, outFlag bool, errFlag, workdir string, args []string) error {
	proc := exec.Command(args[0], args[1:]...)
	proc.Dir = workdir
	logger.Printf("Command path: %v\n", proc.Path)

	const stdoutMarker = 0x00
	const stderrMarker = 0x01

	done := make(chan bool)
	done_count := 0
	done_count += wrapStdin(proc, os.Stdin, inFlag, done)
	if outFlag {
		done_count += wrapStdout(proc, os.Stdout, stdoutMarker, done)
	}
	switch errFlag {
	case "out":
		if outFlag {
			done_count += wrapStderr(proc, os.Stdout, stdoutMarker, done)
		}
	case "err":
		done_count += wrapStderr(proc, os.Stdout, stderrMarker, done)
	case "nil":
		// no-op
	default:
		logger.Panicf("undefined redirect: '%v'\n", errFlag)
	}

	err := proc.Run()
	if e, ok := err.(*exec.Error); ok {
		logger.Printf("Run ERROR: %v\n", e)
		os.Exit(3)
	}
	logger.Printf("Run FINISHED: %#v\n", err)
	for i := 0; i < done_count; i++ {
		<-done
	}
	return err
}
