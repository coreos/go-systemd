// Activation example used by the activation unit tests.
package main

import (
	"fmt"
	"os"

	"github.com/coreos/go-systemd/activation"
)

func fixListenPid() {
	if os.Getenv("FIX_LISTEN_PID") != "" {
		// HACK: real systemd would set LISTEN_PID before exec'ing but
		// this is too difficult in golang for the purpose of a test.
		// Do not do this in real code.
		os.Setenv("LISTEN_PID", fmt.Sprintf("%d", os.Getpid()))
	}
}

func main() {
	fixListenPid()

	files := activation.Files()

	if len(files) == 0 {
		panic("No files")
	}

	// Write out the expected strings to the two pipes
	files[0].Write([]byte("Hello world"))
	files[1].Write([]byte("Goodbye world"))

	return
}
