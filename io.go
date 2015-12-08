package main

import (
	"io"
	"os"
	"os/exec"
)

func wrapStdin(proc *exec.Cmd, stdin io.Reader, done chan bool) {
	logger.Println("Wrapping stdin")

	pipe, err := proc.StdinPipe()
	fatal_if(err)

	go inLoop(pipe, stdin, done)
}

// for protocol v2.0
func wrapStdin2(proc *exec.Cmd, stdin io.Reader, done chan bool) {
	logger.Println("Wrapping stdin")

	pipe, err := proc.StdinPipe()
	fatal_if(err)

	go inLoop2(pipe, proc, stdin, done)
}

func wrapStdout(proc *exec.Cmd, outstream io.Writer, opt byte, done chan bool) {
	logger.Printf("Wrapping stdout with %v\n", opt)

	pipe, err := proc.StdoutPipe()
	fatal_if(err)

	go outLoop(pipe, outstream, opt, done)
}

func wrapStderr(proc *exec.Cmd, outstream io.Writer, opt byte, done chan bool) {
	logger.Printf("Wrapping stderr with %v\n", opt)

	pipe, err := proc.StderrPipe()
	fatal_if(err)

	go outLoop(pipe, outstream, opt, done)
}

///

func inLoop(pipe io.WriteCloser, stdin io.Reader, done chan bool) {
	buf := make([]byte, 2)
	logger.Println("Entering stdin loop")
	done <- true
	for {
		bytes_read, read_err := io.ReadFull(stdin, buf)
		if read_err == io.EOF && bytes_read == 0 {
			break
		}
		fatal_if(read_err)

		length := read16_be(buf)
		logger.Printf("in: packet length = %v\n", length)
		if length == 0 {
			// this is how Porcelain signals EOF from Elixir
			break
		}

		bytes_written, write_err := io.CopyN(pipe, stdin, int64(length))
		logger.Printf("in: copied %v bytes\n", bytes_written)
		fatal_if(write_err)
	}
	pipe.Close()
	done <- true
}

func inLoop2(pipe io.WriteCloser, proc *exec.Cmd, stdin io.Reader, done chan bool) {
	buf := make([]byte, 3)
	logger.Println("Entering stdin loop")
	loop: for {
		bytes_read, read_err := io.ReadFull(stdin, buf[:2])
		if read_err == io.EOF && bytes_read == 0 {
			break
		}
		fatal_if(read_err)

		length := read16_be(buf[:2])
		logger.Printf("in: packet length = %v\n", length)
		if length == 0 {
			// this is how Porcelain signals EOF from Elixir
                        pipe.Close()
                        continue
		}

		_, read_err = io.ReadFull(stdin, buf[2:])
		fatal_if(read_err)

		data_type := buf[2]
		switch data_type {
		case 0:  // input data
			bytes_written, write_err := io.CopyN(pipe, stdin, int64(length)-1)
			logger.Printf("in: copied %v bytes\n", bytes_written)
			fatal_if(write_err)

		case 1:  // signal
			bytes_read, read_err = io.ReadFull(stdin, buf[2:])
			fatal_if(read_err)

			sig := buf[2]
			switch sig {
			case 128:
				sig_err := proc.Process.Signal(os.Interrupt)
				fatal_if(sig_err)
			case 129:
				sig_err := proc.Process.Signal(os.Kill)
				fatal_if(sig_err)
			default:
				sig_err := proc.Process.Signal(makeSignal(sig))
				fatal_if(sig_err)
			}
			break loop

		default:
			logger.Panicf("unhandled input marker: '%v'\n", buf[2])
		}
	}
	pipe.Close()
        logger.Println("Exiting stdin loop")
}

///

// Maximum buffer size for protocol 1.0 is 2 + 2^16-1 - 1
//
//   * 2 is the packet length
//   * 2^16-1 is the maximum amount of data that can be encoded in a
//     2-byte-length packet
//   * 1 byte is used for framing, so it has to be included in the total length
//
var outBuf [1<<16]byte

func outLoop(pipe io.ReadCloser, outstream io.Writer, char byte, done chan bool) {
	buf := outBuf
	buf[2] = char
	logger.Printf("Entering out loop with %v\n", char)
	done <- true
	for {
		bytes_read, read_err := pipe.Read(buf[3:])
		logger.Printf("out: read bytes: %v\n", bytes_read)
		if bytes_read > 0 {
			write16_be(buf[:2], bytes_read+1)
			bytes_written, write_err := outstream.Write(buf[:2+bytes_read+1])
			logger.Printf("out: written bytes: %v\n", bytes_written)
			fatal_if(write_err)
		}
		if read_err == io.EOF || bytes_read == 0 {
			// From io.Reader docs:
			//
			//   Implementations of Read are discouraged from returning a zero
			//   byte count with a nil error, and callers should treat that
			//   situation as a no-op.
			//
			// In this case it appears that 0 bytes may sometimes be returned
			// indefinitely. Therefore we close the pipe.
			if read_err == io.EOF {
				logger.Println("Encountered EOF when reading from stdout")
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

// Unpack the 2-byte integer stored in big endian order
func read16_be(data []byte) uint16 {
	return uint16(data[0]) << 8 | uint16(data[1])
}

// Pack a 2-byte integer in big endian order
func write16_be(data []byte, num int) {
	data[0] = byte(num >> 8)
	data[1] = byte(num)
}
