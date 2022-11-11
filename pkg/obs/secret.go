package obs

import (
	"fmt"
	"os"
)

// Secret is set by the linker flags in the Makefile.
var Secret string

func PrintVersionAndExit1() {
	fmt.Printf("%s\n", Secret)
	os.Exit(0)
}
