package main

import (
	"fmt"
	"os"
)

func die(reason string) {
	if logger != nil {
		logger.Printf("dying: %v\n", reason)
	}
	fmt.Fprintln(os.Stderr, reason)
	os.Exit(-1)
}

func die_usage(reason string) {
	if logger != nil {
		logger.Printf("dying: %v\n", reason)
	}
	fmt.Fprintf(os.Stderr, "%v\n%v\n", reason, usage)
	os.Exit(-1)
}

func fatal(any interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "%v\n", any)
		os.Exit(-1)
	}
	logger.Panicf("%v\n", any)
}

func fatal_if(any interface{}) {
	if logger == nil {
		fmt.Fprintf(os.Stderr, "%v\n", any)
		os.Exit(-1)
	}
	if any != nil {
		logger.Panicf("%v\n", any)
	}
}

/*func shplit(str string) []string {*/
	/*// FIXME*/
	/*return []string{str}*/
/*}*/
