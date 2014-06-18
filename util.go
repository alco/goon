package main

import (
	"fmt"
	"os"
)

func die(reason string) {
	logger.Println("dying:", reason)
	fmt.Fprintln(os.Stderr, reason)
	os.Exit(-1)
}

func die_usage(reason string) {
	logger.Println("dying:", reason)
	fmt.Fprintf(os.Stderr, "%v\n%v\n", reason, usage)
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

/*func shplit(str string) []string {*/
	/*// FIXME*/
	/*return []string{str}*/
/*}*/
