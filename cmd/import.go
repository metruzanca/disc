package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/metruzanca/disc/internal/util"
)

// planAction is a single intended change discovered by the import diff.
type planAction struct {
	action string // "create", "update", or "delete"
	label  string // human-readable description of the affected resource
}

// applyPlan prints the discovered changes grouped by action, a color-coded
// summary of how many create/update/delete changes are planned, and — unless
// dry is set or every action was already known-correct — asks for confirmation
// before returning true. If the plan contains deletes, a red warning is shown
// because deletion is non-reversible.
func applyPlan(path string, changes []planAction, yes, dry bool) bool {
	if len(changes) == 0 {
		util.Yellow.Println("No changes needed.")
		return false
	}

	util.Bold.Printf("Dry run for %s:\n", path)
	fmt.Println()

	hasDelete := false
	var creates, updates, deletes int
	for _, c := range changes {
		switch c.action {
		case "create":
			util.Green.Printf("  create  %s\n", c.label)
			creates++
		case "update":
			util.Yellow.Printf("  update  %s\n", c.label)
			updates++
		case "delete":
			util.Red.Printf("  delete  %s\n", c.label)
			deletes++
			hasDelete = true
		}
	}

	fmt.Println()
	fmt.Printf("  Summary: ")
	util.Green.Printf("%d to create", creates)
	fmt.Printf(", ")
	util.Yellow.Printf("%d to update", updates)
	fmt.Printf(", ")
	util.Red.Printf("%d to delete", deletes)
	fmt.Println()

	if hasDelete {
		fmt.Println()
		util.Red.Println("WARNING: Deleting is a non-reversible action.")
	}

	fmt.Println()
	if dry {
		return false
	}
	return util.ConfirmRun("Apply these changes?", yes, false)
}

// readJSONFile reads and unmarshals a JSON file into out.
func readJSONFile(path string, out interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return nil
}
