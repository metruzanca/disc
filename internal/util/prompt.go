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

// Confirm asks the user a yes/no question and returns true if they agree.
// If yes is true, it skips the prompt and returns true immediately.
func Confirm(prompt string, yes bool) bool {
	if yes {
		return true
	}
	fmt.Printf("%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}
