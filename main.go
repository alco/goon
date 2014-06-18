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

func fatal_if(any interface{}) {
	if any != nil {
		logger.Panicln(any)
	}
}

func shplit(str string) []string {
	// FIXME
	return []string{str}
}

// Unpack the 2-byte integer stored in big endian order
func read16_be(data []byte) int16 {
	return int16(data[0]) << 8 | int16(data[1])
}

// Pack a 2-byte integer in big endian order
func write16_be(data []byte, num int) {
	data[0] = byte(num >> 8)
	data[1] = byte(num)
}

func inLoop(pipe io.WriteCloser, stdin io.Reader, done chan bool) {
	buf := make([]byte, 2)
	logger.Println("Entering stdin loop")
	for {
		bytes_read, read_err := io.ReadFull(stdin, buf)
		if read_err == io.EOF && bytes_read == 0 {
			pipe.Close()
			break
		}
		fatal_if(read_err)

		length := read16_be(buf)
		logger.Printf("in: packet length = %v\n", length)
		if length == 0 {
			// this is how Porcelain signals EOF from Elixir
			pipe.Close()
			break
		}

		bytes_written, write_err := io.CopyN(pipe, stdin, int64(length))
		logger.Printf("in: copied %v bytes\n", bytes_written)
		fatal_if(write_err)
	}
	done <- true
}

func wrapStdin(proc *exec.Cmd, stdin io.Reader, inFlag bool, done chan bool) int {
	if !inFlag {
		return 0
	}

	/*fmt.Fprintf(os.Stderr, "Wrapping stdin")*/

	pipe, err := proc.StdinPipe()
	fatal_if(err)

	go inLoop(pipe, stdin, done)
	return 1
}

func outLoop(pipe io.ReadCloser, outstream io.Writer, char byte, done chan bool) {
	buf := make([]byte, kOutputBufferSize)
	buf[2] = char
	logger.Printf("Entering out loop with %v\n", char)
	for {
		bytes_read, read_err := pipe.Read(buf[3:])
		logger.Printf("out: read bytes: %v\n", bytes_read)
		if bytes_read > 0 {
			write16_be(buf[:2], bytes_read+1)
			bytes_written, write_err := outstream.Write(buf[:2+bytes_read+1])
			logger.Printf("out: written bytes: %v\n", bytes_written)
			fatal_if(write_err)
		}
		if read_err == io.EOF /*|| bytes_read == 0*/ {
			// !!!
			// The note below is currently irrelevant, but left here in case
			// the bug reappers in the future.
			// !!!

			// From io.Reader docs:
			//
			//   Implementations of Read are discouraged from returning a zero
			//   byte count with a nil error, and callers should treat that
			//   situation as a no-op.
			//
			// In this case it appears that 0 bytes may be returned
			// indefinitely when reading from stderr. Therefore we close the pipe.
			if read_err == io.EOF {
				logger.Println("Encountered EOF on input")
			} else {
				logger.Println("Read 0 bytes with no error")
			}
			break
		}
		if read_err != nil {
			switch read_err.(type) {
			case *os.PathError:
				// known error
				break
			default:
				fatal(read_err)
			}
		}
	}
	pipe.Close()
	done <- true
}

func wrapStdout(proc *exec.Cmd, outstream io.Writer, opt byte, done chan bool) int {
	pipe, err := proc.StdoutPipe()
	fatal_if(err)

	/*fmt.Fprintf(os.Stderr, "Wrapping stdout with %v\n", opt)*/

	go outLoop(pipe, outstream, opt, done)
	return 1
}

func wrapStderr(proc *exec.Cmd, outstream io.Writer, opt byte, done chan bool) int {
	pipe, err := proc.StderrPipe()
	fatal_if(err)

	/*fmt.Fprintf(os.Stderr, "Wrapping stderr with %v\n", opt)*/

	go outLoop(pipe, outstream, opt, done)
	return 1
}

var protoFlag = flag.String("proto", "", "protocol version (one of: 0.0)")
var inFlag  = flag.Bool("in", false, "whether stdin is used")
var outFlag = flag.Bool("out", false, "whether stdout is preserved or discarded")
var errFlag = flag.String("err", "nil", "redirection or supression of stderr")
var dirFlag = flag.String("dir", ".", "working directory for the spawned process")

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
	var protoImpl func(bool, bool, string, string, []string) error
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
