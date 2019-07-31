// Debug/diagnostic functions
package debug

import "fmt"

// Prints formatted text in yellow
func PrintfLnCaption(format string, args ...interface{}) {
	fmt.Printf("\x1b[1;33m"+format+"\x1b[0m\n", args...)
}

// Prints an array of strings
func PrintTextArray(text []string) {
	for _, line := range text {
		fmt.Printf("%s", line)
	}
	fmt.Printf("\n")
}

func DbgPrint(text string) {
	fmt.Print("\x1B[1;31mDEBUG: \x1B[0m" + text)
}

func DbgPrintln(text string) {
	fmt.Print("\x1B[1;31mDEBUG: \x1B[0m" + text + "\n")
}

func DbgPrintf(format string, args ...interface{}) {
	fmt.Printf("\x1B[1;31mDEBUG: \x1B[0m"+format, args...)
}

// Prints the specified number of elements from an array of strings and indicates
// how many more elements there are
func PrintTextArrayLimited(text []string, limit int) {
	lineCount := len(text)
	for idx := 0; idx < lineCount && idx < limit; idx++ {
		line := text[idx]
		fmt.Printf("%s", line)
	}
	if limit < lineCount {
		fmt.Printf("... %d more lines ...\n", lineCount-limit)
	}
	fmt.Printf("\n")
}
