// Package ui provides terminal form helpers built on charmbracelet/huh.
// Forms are only meant to run in an interactive terminal; agents and scripts
// should keep stdin non-interactive and use flags instead.
package ui

import (
	"os"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

// IsTerminal reports whether standard input is an interactive terminal.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// ChannelOption is a selectable channel for hosting events.
type ChannelOption struct {
	Name  string
	ID    string
	Voice bool
}

// conditional overrides a field's Skip() so it only shows when show() is true.
type conditional struct {
	huh.Field
	show func() bool
}

func (c conditional) Skip() bool {
	return !c.show()
}

// onlyWhen wraps f so it is only rendered when cond is true.
func onlyWhen(f huh.Field, cond func() bool) huh.Field {
	return conditional{Field: f, show: cond}
}
