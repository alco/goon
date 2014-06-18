package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

var logger *log.Logger
const logsEnabled = false

func init() {
	var filename string
	if logsEnabled {
		filename = "goon.log"
	} else {
		filename = os.DevNull
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	logger = log.New(file, "goon", log.Lmicroseconds)
}

const kOutputBufferSize = 1024

func die(reason string) {
	logger.Println("dying:", reason)
	println(reason)
	os.Exit(-1)
}

func die_usage(reason string) {
	logger.Println("dying:", reason)
	println(reason)
	println(usage)
	os.Exit(-1)
}

func fatal(any interface{}) {
	logger.Panicln(any)
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

func wrapIn(pipe io.WriteCloser, stdin io.Reader, done chan bool) {
	buf := make([]byte, 2)
	logger.Println("Entering stdin loop")
	for {
		nbytes, err := io.ReadFull(stdin, buf)
		if err != nil {
			if err == io.EOF && nbytes == 0 {
				pipe.Close()
				break
			}
			fatal(err)
		}
		/*fmt.Fprintf(os.Stderr, "Read %v bytes\n", nbytes)*/

		length := read16_be(buf)
		logger.Printf("in: packet length = %v\n", length)
		/*fmt.Fprintf(os.Stderr, "length = %v\n", length)*/
		if length == 0 {
			// this is how Porcelain signals EOF from Elixir
			pipe.Close()
			break
		}

		nnbytes, err := io.CopyN(pipe, stdin, int64(length))
		logger.Printf("in: copied %v bytes\n", nnbytes)
		if err != nil {
			fatal(err)
		}
	}
	done <- true
}

func wrapStdin(proc *exec.Cmd, stdin io.Reader, inFlag bool, done chan bool) int {
	if !inFlag {
		return 0
	}

	pipe, err := proc.StdinPipe()
	if err != nil {
		fatal(err)
	}
	go wrapIn(pipe, stdin, done)
	return 1
}

func wrapOut(pipe io.ReadCloser, outstream io.Writer, out string, done chan bool) int {
	var char byte

	if out == "out" {
		char = 'o'
	} else if out == "err" {
		char = 'e'
	} else if out == "nil" {
		return 0
	} else {
		fatal("undefined redirect")
	}

	go func() {
		buf := make([]byte, kOutputBufferSize)
		buf[2] = char
		logger.Printf("Entering out loop with %v\n", out)
		for {
			nbytes, err := pipe.Read(buf[3:])
			logger.Printf("out: read bytes: %v\n", nbytes)
			if nbytes > 0 {
				write16_be(buf[:2], nbytes+1)
				outstream.Write(buf[:2+nbytes+1])
			}

			if err == io.EOF {
				break
			}
			if err != nil {
				switch err.(type) {
				case *os.PathError:
					// known error
					break
				default:
					panic(err)
				}
			}
		}
		done <- true
	}()

	return 1
}

func wrapStdout(proc *exec.Cmd, outstream io.Writer, out string, done chan bool) int {
	pipe, err := proc.StdoutPipe()
	if err != nil {
		fatal(err)
	}
	return wrapOut(pipe, outstream, out, done)
}

func wrapStderr(proc *exec.Cmd, outstream io.Writer, out string, done chan bool) int {
	pipe, err := proc.StderrPipe()
	if err != nil {
		fatal(err)
	}
	return wrapOut(pipe, outstream, out, done)
}

var protoFlag = flag.String("proto", "", "protocol version (one of: 0.0)")
var inFlag  = flag.Bool("in", false, "specify whether stdin will be used")
var outFlag = flag.String("out", "out", "specify redirection or supression of stdout")
var errFlag = flag.String("err", "err", "specify redirection or supression of stderr")
var dirFlag = flag.String("dir", ".", "specify working directory for the spawned process")

const usage = "Usage: goon -proto <version> [options] -- <program> [<arg>...]"

func main() {
	flag.Parse()
	args := flag.Args()

	/* Validate options and arguments */
	if *protoFlag == "" {
		die_usage("Please specify the protocol version.")
	}

	if len(args) < 1 {
		die_usage("Not enough arguments.")
	}

	/* Choose protocol implementation */
	var protoImpl func(bool, string, string, string, []string) error
	switch *protoFlag {
	case "0.0":
		protoImpl = proto_0_0
	default:
		reason := fmt.Sprintf("Unsupported protocol version: %v", *protoFlag)
		die(reason)
	}

	/* Run external program and block until it terminates */
	err := protoImpl(*inFlag, *outFlag, *errFlag, *dirFlag, args)

	/* Determine the exit status */
	if err != nil {
		//fmt.Printf("%#v\n", err)
		os.Exit(get_exit_status(err))
	}
}
