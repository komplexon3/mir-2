package utils

import (
	"fmt"

	"github.com/filecoin-project/mir/fuzzer/checker"
)

func PrintResult(label string, result checker.CheckerResult) string {
	return fmt.Sprintf("%s: %s", label, result.String())
}
