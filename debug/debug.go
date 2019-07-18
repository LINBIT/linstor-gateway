package debug

import "fmt"

func PrintfLnCaption(format string, args ...interface{}) {
    fmt.Printf("\x1b[1;33m" + format + "\x1b[0m\n", args...)
}

func PrintTextArray(text []string) {
	for _, line := range text {
		fmt.Printf("%s", line)
	}
	fmt.Printf("\n")
}

func PrintTextArrayLimited(text []string, limit int) {
	lineCount := len(text)
	for idx := 0; idx < lineCount && idx < limit; idx++ {
		line := text[idx]
		fmt.Printf("%s", line)
	}
	if (limit < lineCount) {
		fmt.Printf("... %d more lines ...\n", lineCount - limit)
	}
	fmt.Printf("\n")
}
