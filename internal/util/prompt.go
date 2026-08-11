// Package util provides shared helpers for prompts and output styling.
package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

var (
	// Bold is used for names/headings.
	Bold = color.New(color.Bold)
	// Dim is used for secondary text such as IDs.
	Dim = color.New(color.Faint)
	// Green is used for success messages.
	Green = color.New(color.FgGreen)
	// Red is used for errors.
	Red = color.New(color.FgRed)
	// Cyan is used for labels.
	Cyan = color.New(color.FgCyan)
	// Yellow is used for warnings.
	Yellow = color.New(color.FgYellow)
)

// ConfirmRun gates a mutating action. In dry mode it prints the intended
// action and returns false without prompting or acting. In yes mode it
// returns true immediately. Otherwise it asks for y/N, printing a message
// when it does not proceed. dry takes priority over yes.
func ConfirmRun(summary string, yes, dry bool) bool {
	if dry {
		Cyan.Printf("DRY RUN: %s\n", summary)
		return false
	}
	if yes {
		return true
	}
	fmt.Printf("%s [y/N] ", summary)
	line := readLine()
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "y" || line == "yes" {
		return true
	}
	Yellow.Println("Aborted.")
	return false
}

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimRight(line, "\n")
}
