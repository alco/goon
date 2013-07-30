package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
)

/*func spawn(cmdname string, args []string) {*/
	/*cmd := exec.Command(cmdname, args...)*/
/*}*/

// read *all* input up to the newline
func readln(r *bufio.Reader) ([]byte, error) {
    var (
        isPrefix bool = true
        err error     = nil
        line, ln []byte
    )

    for isPrefix && err == nil {
        line, isPrefix, err = r.ReadLine()
        ln = append(ln, line...)
    }

    return ln, err
}

type Process struct {
	cmd *exec.Cmd
	pin io.WriteCloser
	pout io.ReadCloser
}

func NewProcess(cmdname string, args... string) (p *Process, err error) {
	cmd := exec.Command(cmdname, args...)

	pin, err := cmd.StdinPipe()
	if err != nil {
		return
	}

	pout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	p = &Process{cmd: cmd, pin: pin, pout: pout}
	return
}

func (p *Process) Write(data []byte) (int, error) {
	return p.pin.Write(data)
}

func (p *Process) WriteFrom(r io.Reader) error {
	// ...
	return nil
}

func (p *Process) SendEOF() error {
	return p.pin.Close()
}

func (p *Process) Read() ([]byte, error) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(p.pout)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Process) ReadInto(buf []byte) (int, error) {
	nbytes, err := p.pout.Read(buf)
	return nbytes, err
}

func (p *Process) Execute(input string) (string, error) {
	err := p.cmd.Start()
	if err != nil {
		return "", err
	}

	go func() {
		p.Write([]byte(input))
		p.SendEOF()
	}()

	data, err := p.Read()
	return string(data), err
}

func main() {
	proc, err := NewProcess("/Users/alco/Downloads/Pygments-1.6/pygmentize", "-f", "html", "-l", "javascript")
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(os.Stdin)

	output, err := proc.Execute(string(buf.Bytes()))
	if err != nil {
		panic(err)
	}
	print(output)
}

func mmain() {
	cmd := exec.Command("/Users/alco/Downloads/Pygments-1.6/pygmentize", "-f", "html", "-l", "javascript")

	pipe_in, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	pipe_out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	done := make(chan bool)

	go func() {
		r := bufio.NewReader(os.Stdin)
		for {
			line, eof := readln(r)
			fmt.Println("Got line", line)
			pipe_in.Write(line)
			if eof != nil {
				println("Got eof")
				err := pipe_in.Close()
				if err != nil {
					panic(err)
				}
				break;
			}
		}
	}()

	go func() {
		var buf bytes.Buffer
		_, err := buf.ReadFrom(pipe_out)
		if err != nil {
			panic(err)
		}
		print(string(buf.Bytes()))
		done <- true
	}()

	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	<-done
}
