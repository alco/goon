package main

import (
	"os/exec"
	"syscall"
)

func getExitStatus(err error) int {
	switch e := err.(type) {
	case *exec.ExitError:
		switch s := e.ProcessState.Sys().(type) {
		case syscall.Waitmsg:
			return s.ExitStatus()
		}
	}
	return 1
}
