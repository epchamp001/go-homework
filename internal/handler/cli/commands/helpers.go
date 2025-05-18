package commands

import (
	"fmt"
	"os"
)

func printErr(err error) {
	fmt.Fprintln(os.Stderr, "ERROR:", err)
}
