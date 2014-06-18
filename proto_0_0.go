package main

import (
	"os"
	"os/exec"
)

func proto_0_0(inFlag, outFlag bool, errFlag, workdir string, args []string) error {
	const stdoutMarker = 0x00
	const stderrMarker = 0x01

	proc := exec.Command(args[0], args[1:]...)
	proc.Dir = workdir

	logger.Printf("Command path: %v\n", proc.Path)

	done := make(chan bool)
	doneCount := 0

	if inFlag {
		wrapStdin(proc, os.Stdin, done)
		doneCount++
	}

	if outFlag {
		wrapStdout(proc, os.Stdout, stdoutMarker, done)
		doneCount++
	}

	switch errFlag {
	case "out":
		if outFlag {
			wrapStderr(proc, os.Stdout, stdoutMarker, done)
			doneCount++
		}
	case "err":
		wrapStderr(proc, os.Stdout, stderrMarker, done)
		doneCount++
	case "nil":
		// no-op
	default:
		logger.Panicf("undefined redirect: '%v'\n", errFlag)
	}

	// FIXME: perform proper handshake instead
	err := proc.Run()
	if e, ok := err.(*exec.Error); ok {
		// This shouldn't really happen in practice because we check for
		// program existence in Elixir, before launching goon
		logger.Printf("Run ERROR: %v\n", e)
		os.Exit(3)
	}
	logger.Printf("Run FINISHED: %#v\n", err)
	for i := 0; i < doneCount; i++ {
		<-done
	}
	return err
}
