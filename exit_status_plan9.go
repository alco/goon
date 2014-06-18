package main

import (
	"os/exec"
	"syscall"
)

func get_exit_status(err error) int {
	switch e := err.(type) {
	case *exec.ExitError:
		switch s := e.ProcessState.Sys().(type) {
		case syscall.Waitmsg:
			return s.ExitStatus()
		}
	}
	return 1
}
