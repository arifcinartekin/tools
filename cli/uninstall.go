package main

import (
	"fmt"

	"github.com/arifcinartekin/tools/cli/internal/installer"
	"github.com/arifcinartekin/tools/cli/internal/state"
)

func runUninstall(names []string) error {
	if len(names) == 0 {
		return fmt.Errorf("uninstall requires at least one item name")
	}

	st, err := state.Load()
	if err != nil {
		return err
	}

	for _, name := range names {
		entry, ok := st.Items[name]
		if !ok {
			return fmt.Errorf("%q is not installed", name)
		}

		if err := installer.Uninstall(entry.TargetPath, entry.LinkPath, entry.Kind); err != nil {
			return fmt.Errorf("uninstalling %s: %w", name, err)
		}

		st.Remove(name)
		if err := st.Save(); err != nil {
			return err
		}
		fmt.Printf("Uninstalled %s\n", name)
	}
	return nil
}
