package main

import (
	"fmt"
	"os"
	"os/exec"
)

func proto_0_0(inFlag, outFlag bool, errFlag, workdir string, args []string) error {
	proc := exec.Command(args[0], args[1:]...)
	proc.Dir = workdir

	done := make(chan bool)
	done_count := 0
	done_count += wrapStdin(proc, os.Stdin, inFlag, done)
	if outFlag {
		done_count += wrapStdout(proc, os.Stdout, 'o', done)
	}
	if errFlag == "out" && outFlag {
		done_count += wrapStderr(proc, os.Stdout, 'o', done)
	} else if errFlag == "err" {
		done_count += wrapStderr(proc, os.Stdout, 'e', done)
	} else if errFlag != "nil" {
		fmt.Fprintf(os.Stderr, "undefined redirect: '%v'\n", errFlag)
		fatal("undefined redirect")
	}

	err := proc.Run()
	for i := 0; i < done_count; i++ {
		<-done
	}
	return err
}
