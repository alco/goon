package main

import (
	"os"
	"os/exec"
)

func proto_0_0(inFlag bool, outFlag, errFlag, workdir string, args []string) error {
	proc := exec.Command(args[0], args[1:]...)
	proc.Dir = workdir

	done := make(chan bool)
	done_count := 0
	done_count += wrapStdin(proc, os.Stdin, inFlag, done)
	if opt := checkOutOpt(outFlag); opt != 0 {
		done_count += wrapStdout(proc, os.Stdout, outFlag, done)
	}
	if opt := checkOutOpt(errFlag); opt != 0 {
		done_count += wrapStderr(proc, os.Stdout, errFlag, done)
	}

	err := proc.Run()
	for i := 0; i < done_count; i++ {
		<-done
	}
	return err
}
