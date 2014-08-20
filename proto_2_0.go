package main

import (
	"os"
	"os/exec"
)

func proto_2_0(inFlag, outFlag bool, errFlag, workdir string, args []string) error {
	const stdoutMarker = 0x00
	const stderrMarker = 0x01

	proc := exec.Command(args[0], args[1:]...)
	proc.Dir = workdir

	logger.Printf("Command path: %v\n", proc.Path)

	// The channel is buffered to reduce waiting time when exchanging acks
	// between the main routine and io routines.
	doneChan := make(chan bool, 3)
	doneCount := 0

	if inFlag {
		wrapStdin2(proc, os.Stdin, doneChan)
		doneCount++
	}

	if outFlag {
		wrapStdout(proc, os.Stdout, stdoutMarker, doneChan)
		doneCount++
	}

	switch errFlag {
	case "out":
		if outFlag {
			wrapStderr(proc, os.Stdout, stdoutMarker, doneChan)
			doneCount++
		}
	case "err":
		wrapStderr(proc, os.Stdout, stderrMarker, doneChan)
		doneCount++
	case "nil":
		// no-op
	default:
		logger.Panicf("undefined redirect: '%v'\n", errFlag)
	}

	// Initial ack to make sure all pipes have been connected
	for i := 0; i < doneCount; i++ {
		<-doneChan
	}

	// FIXME: perform proper handshake instead
	err := proc.Start()
	fatal_if(err)

	// Finishing ack to ensure all pipes were closed
	for i := 0; i < doneCount; i++ {
		<-doneChan
	}

	err = proc.Wait()
	if e, ok := err.(*exec.Error); ok {
		// This shouldn't really happen in practice because we check for
		// program existence in Elixir, before launching goon
		logger.Printf("Run ERROR: %v\n", e)
		os.Exit(3)
	}
	logger.Printf("Run FINISHED: %#v\n", err)
	return err
}
