package main

import (
	"os"
)

func main() {
	buf := make([]byte, 10)
	buf[0] = 0
	buf[1] = 5
	os.Stdout.Write(buf[:2])

	copy(buf, "hello")
	os.Stdout.Write(buf)

	buf[0] = 0
	buf[1] = 0
	os.Stdout.Write(buf[:2])
}
