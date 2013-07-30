package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

const kOutputBufferSize = 1 << 10

func die(any interface{}) {
	fmt.Printf("%v\n", any);
	os.Exit(1)
}

func fatal(any interface{}) {
	panic(any)
}

func read16_be(data []byte) int16 {
	return int16(data[0]) << 8 | int16(data[1])
}

func write16_be(data []byte, num int) {
	data[0] = byte(num >> 8)
	data[1] = byte(num)
}

func wrapStdin(proc *exec.Cmd, stdin io.Reader) {
	pipe, err := proc.StdinPipe()
	if err != nil {
		fatal(err)
	}

	go func() {
		buf := make([]byte, 2)
		for {
			nbytes, err := io.ReadFull(stdin, buf)
			if err != nil {
				if err == io.EOF && nbytes == 0 {
					pipe.Close()
					break
				}
				fatal(err)
			}

			length := read16_be(buf)
			if length == 0 {
				// EOF
				pipe.Close()
				break
			}

			_, err = io.CopyN(pipe, stdin, int64(length))
			if err != nil {
				fatal(err)
			}
		}
	}()
}

func wrapStdout(proc *exec.Cmd, stdout io.Writer) {
	pipe, err := proc.StdoutPipe()
	if err != nil {
		fatal(err)
	}

	go func() {
		buf := make([]byte, kOutputBufferSize)
		for {
			nbytes, err := pipe.Read(buf[2:])
			if nbytes > 0 {
				write16_be(buf[:2], nbytes)
				stdout.Write(buf[:2+nbytes])
			}

			if err == io.EOF {
				break
			}
			if err != nil {
				fatal(err)
			}
		}
	}()
}

func main() {
	// First, we see which program needs to be launched
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		die("Not enough arguments.\nSynopsis: goon <program> [arg1] ...")
	}

	proc := exec.Command(args[0], args[1:]...)
	wrapStdin(proc, os.Stdin)
	wrapStdout(proc, os.Stdout)

	// Now we're ready to start the requested program
	err := proc.Run()
	if err != nil {
		fatal(err)
	}
}
