package progress

import (
	"fmt"
	"io"
	"os"
)

// Writer is the destination for progress output. Tests can replace it to capture output.
var Writer io.Writer = os.Stdout

const totalStages = 5

// StageStart prints a stage start line in [N/5] format with trailing ellipsis.
func StageStart(stage int, msg string) {
	fmt.Fprintf(Writer, "[%d/%d] %s...\n", stage, totalStages, msg)
}

// StageComplete prints a stage completion line in [N/5] format.
func StageComplete(stage int, msg string) {
	fmt.Fprintf(Writer, "[%d/%d] %s\n", stage, totalStages, msg)
}
