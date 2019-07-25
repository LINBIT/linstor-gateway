package application

import "fmt"

const (
	COLOR_BLACK      = "\x1B[0;30m"
	COLOR_DARK_RED   = "\x1B[0;31m"
	COLOR_DARK_GREEN = "\x1B[0;32m"
	COLOR_BROWN      = "\x1B[0;33m"
	COLOR_DARK_BLUE  = "\x1B[0;34m"
	COLOR_TEAL       = "\x1B[0;35m"
	COLOR_DARK_PINK  = "\x1B[0;36m"
	COLOR_GRAY       = "\x1B[0;37m"
	COLOR_DARK_GRAY  = "\x1B[1;30m"
	COLOR_RED        = "\x1B[1;31m"
	COLOR_GREEN      = "\x1B[1;32m"
	COLOR_YELLOW     = "\x1B[1;33m"
	COLOR_BLUE       = "\x1B[1;34m"
	COLOR_CYAN       = "\x1B[1;35m"
	COLOR_PINK       = "\x1B[1;36m"
	COLOR_WHITE      = "\x1B[1;37m"
	COLOR_RESET      = "\x1B[0m"
)

func color(clrFmt string) {
	fmt.Print(clrFmt)
}

func defaultColor() {
	fmt.Print(COLOR_RESET)
}
