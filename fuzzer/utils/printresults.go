package utils

import (
	"fmt"

	"github.com/filecoin-project/mir/fuzzer/checker"
)

func FormatResult(label string, result checker.CheckerResult) string {
	return fmt.Sprintf("%s: %s", label, result.String())
}
