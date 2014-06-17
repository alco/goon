package main

import (
	"os"
	"os/exec"
)

func proto_0_0(inFlag, outFlag, errFlag, workdir string, args []string) error {
	done := make(chan bool)
	done_count := 0

	proc := exec.Command(args[0], args[1:]...)
	proc.Dir = workdir
	if inFlag != "nil" {
		done_count += wrapStdin(proc, os.Stdin, done)
	}
	/*if shouldWrapOut(*outFlag, *errFlag, "out", "err") {*/
		done_count += wrapStdout(proc, os.Stdout, outFlag, done)
	/*}*/
	/*if shouldWrapOut(*errFlag, *outFlag, "err", "out") {*/
		done_count += wrapStderr(proc, os.Stdout, os.Stderr, errFlag, done)
	/*}*/

	// Now we're ready to start the requested program
	err := proc.Run()
	for i := 0; i < done_count; i++ {
		<-done
	}

	return err
}
