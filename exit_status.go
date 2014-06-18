// +build !plan9

package main

import (
	"os/exec"
	"syscall"
)

func get_exit_status(err error) int {
	switch e := err.(type) {
	case *exec.ExitError:
		switch s := e.ProcessState.Sys().(type) {
		case syscall.WaitStatus:
			return s.ExitStatus()
		}
	}
	return 1
}
