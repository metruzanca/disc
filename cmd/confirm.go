package cmd

import (
	"fmt"

	"github.com/metruzanca/disc/internal/ui"
	"github.com/metruzanca/disc/internal/util"
)

// interactive reports whether interactive forms should be shown. They only
// appear in a real terminal, and only when the caller has not opted out via
// --yes/--dry or --agent.
func interactive(yes, dry bool) bool {
	return !agentMode && !yes && !dry && ui.IsTerminal()
}

// confirmRun gates a mutating action. proceed reports whether to continue.
// In --agent mode without --yes/--dry it errors instead of prompting, so the
// CLI is guaranteed to never block on stdin.
func confirmRun(summary string, yes, dry bool) (proceed bool, err error) {
	if agentMode && !yes && !dry {
		return false, fmt.Errorf("action requires confirmation; pass --yes to apply or --dry to preview (prompts are disabled by --agent)")
	}
	return util.ConfirmRun(summary, yes, dry), nil
}
