package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

const kOutputBufferSize = 1024

const (
	kProtocolOneToOne = iota
	kProtocol2Length
)

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

func stdin_2l(pipe io.WriteCloser, stdin io.Reader, done chan bool) {
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
}

func wrapStdin(proc *exec.Cmd, stdin io.Reader, done chan bool, proto int) int {
	if proto == kProtocolOneToOne {
		proc.Stdin = stdin
		return 0
	}

	if proto == kProtocol2Length {
		pipe, err := proc.StdinPipe()
		if err != nil {
			fatal(err)
		}
		go stdin_2l(pipe, stdin, done)
		return 1
	}

	fatal("Unknown protocol")
	return 0
}

func wrapOut(pipe io.ReadCloser, outstream io.Writer, out string, done chan bool) int {
	var char byte

	if out == "out" {
		char = 'o'
	} else if out == "err" {
		char = 'e'
	} else {
		fatal("undefined redirect")
	}

	go func() {
		buf := make([]byte, kOutputBufferSize)
		buf[2] = char
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

	return 1
}

func wrapStdout(proc *exec.Cmd, outstream io.Writer, errstream io.Writer, out string, done chan bool, proto int) int {
	if proto == kProtocolOneToOne {
		if out == "out" {
			proc.Stdout = outstream
		} else if out == "err" {
			proc.Stdout = errstream
		} else {
			fatal("undefined redirect")
		}
		return 0
	}

	if proto == kProtocol2Length {
		pipe, err := proc.StdoutPipe()
		if err != nil {
			fatal(err)
		}
		return wrapOut(pipe, outstream, out, done)
	}

	fatal("Unknown protocol")
	return 0
}

func wrapStderr(proc *exec.Cmd, outstream io.Writer, errstream io.Writer, out string, done chan bool, proto int) int {
	if proto == kProtocolOneToOne {
		if out == "out" {
			proc.Stderr = outstream
		} else if out == "err" {
			proc.Stderr = errstream
		} else {
			fatal("undefined redirect")
		}
		return 0
	}

	if proto == kProtocol2Length {
		pipe, err := proc.StderrPipe()
		if err != nil {
			fatal(err)
		}
		return wrapOut(pipe, outstream, out, done)
	}

	fatal("Unknown protocol")
	return 0
}

func shouldWrapOut(out string, err string, opt_out string, opt_err string) bool {
	result := (out == opt_out)
	if out == opt_err {
		if err == opt_err {
			result = true
		} else if len(err) != 0 {
			fatal("Invalid redirection spec")
		}
	}
	return result
}

var protoFlag = flag.String("proto", "0.0", "protocol version (one of: 0.0)")
var outFlag = flag.String("out", "out", "specify redirection or supression of stdout")
var errFlag = flag.String("err", "err", "specify redirection or supression of stderr")

func main() {
	flag.Parse()
	args := flag.Args()

	fmt.Fprintf(os.Stderr, "%#v\n", args)
	fmt.Fprintf(os.Stderr, "%#v\n", *outFlag)
	fmt.Fprintf(os.Stderr, "%#v\n", *errFlag)
	fmt.Fprintf(os.Stderr, "%#v\n", *protoFlag)
	/*return*/

	if len(args) < 1 {
		die("Not enough arguments.\nSynopsis: goon [opts] <program> [<arg>...]")
	}

	var impl func(*string, *string, []string) error
	switch *protoFlag {
	case "0.0":
		impl = proto_0_0
	default:
		die("Unknown protocol")
	}

	err := impl(outFlag, errFlag, args)

	/*done := make(chan bool)*/
	/*done_count := 0*/

	/*proc := exec.Command(args[0], args[1:]...)*/
	/*done_count += wrapStdin(proc, os.Stdin, done, proto)*/
	/*if shouldWrapOut(*outFlag, *errFlag, "out", "err") {*/
		/*done_count += wrapStdout(proc, os.Stdout, os.Stderr, *outFlag, done, proto)*/
	/*}*/
	/*if shouldWrapOut(*errFlag, *outFlag, "err", "out") {*/
		/*done_count += wrapStderr(proc, os.Stdout, os.Stderr, *errFlag, done, proto)*/
	/*}*/

	/*// Now we're ready to start the requested program*/
	/*err := proc.Run()*/
	/*for i := 0; i < done_count; i++ {*/
		/*<-done*/
	/*}*/

	// Determine the exit status
	if err != nil {
		fmt.Printf("%#v\n", err)
		os.Exit(get_exit_status(err))
	}
}
