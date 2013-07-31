package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

const kOutputBufferSize = 1024

func die(any interface{}) {
	fmt.Printf("%v\n", any);
	os.Exit(1)
}

func fatal(any interface{}) {
	panic(any)
}

func shplit(str string) []string {
	// FIXME
	return []string{str}
}

func read16_be(data []byte) int16 {
	return int16(data[0]) << 8 | int16(data[1])
}

func write16_be(data []byte, num int) {
	data[0] = byte(num >> 8)
	data[1] = byte(num)
}

func wrapStdin(proc *exec.Cmd, stdin io.Reader, done chan bool) {
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
		done <- true
	}()
}

func wrapOut(pipe io.ReaderCloser, outstream io.Writer, done chan bool) {
	go func() {
		buf := make([]byte, kOutputBufferSize)
		buf[2] = 'o'
		for {
			nbytes, err := pipe.Read(buf[3:])
			if nbytes > 0 {
				write16_be(buf[:2], nbytes+1)
				outstream.Write(buf[:2+nbytes+1])
			}

			if err == io.EOF {
				break
			}
			if err != nil {
				fatal(err)
			}
		}
		done <- true
	}()
}

func wrapStdout(proc *exec.Cmd, stdout io.Writer, done chan bool) {
	pipe, err := proc.StdoutPipe()
	if err != nil {
		fatal(err)
	}
	wrapOut(pipe, stdout, done)
}

func wrapStderr(proc *exec.Cmd, stderr io.Writer, done chan bool) {
	pipe, err := proc.StderrPipe()
	if err != nil {
		fatal(err)
	}
	wrapOut(pipe, stderr, done)
}

func main() {
	// First, we see which program needs to be launched
	flag.Parse()
	args := flag.Args()
	/*fmt.Printf("%#v\n", args)*/

	if len(args) < 1 {
		die("Not enough arguments.\nSynopsis: goon <program> [arg1] ...")
	}

	if len(args) == 1 {
		// We need to parse the arguments ourselves
		args = shplit(args[0])
	}

	done := make(chan bool)
	proc := exec.Command(args[0], args[1:]...)
	wrapStdin(proc, os.Stdin, done)
	wrapStdout(proc, os.Stdout, done)
	wrapStderr(proc, os.Stdout, done)

	// Now we're ready to start the requested program
	_ = proc.Run()
	<-done
	<-done
	<-done
	/*if err != nil {*/
		/*fatal(err)*/
	/*}*/
}
