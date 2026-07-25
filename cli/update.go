package main

import (
	"fmt"
	"runtime"

	"github.com/arifcinartekin/tools/cli/internal/installer"
	"github.com/arifcinartekin/tools/cli/internal/manifest"
	"github.com/arifcinartekin/tools/cli/internal/semver"
	"github.com/arifcinartekin/tools/cli/internal/state"
)

func runUpdate(args []string) error {
	all := false
	var names []string
	for _, a := range args {
		if a == "--all" {
			all = true
			continue
		}
		names = append(names, a)
	}
	if !all && len(names) == 0 {
		return fmt.Errorf("update requires item name(s) or --all")
	}

	m, err := manifest.Fetch(baseURL())
	if err != nil {
		return err
	}

	st, err := state.Load()
	if err != nil {
		return err
	}

	if all {
		names = names[:0]
		for name := range st.Items {
			names = append(names, name)
		}
		if len(names) == 0 {
			fmt.Println("Nothing installed.")
			return nil
		}
	}

	for _, name := range names {
		entry, installed := st.Items[name]
		if !installed {
			return fmt.Errorf("%q is not installed (use 'toolbox install %s')", name, name)
		}
		it, ok := m.Find(name)
		if !ok {
			return fmt.Errorf("%q no longer exists in the repository", name)
		}
		if !it.SupportsCurrentOS(runtime.GOOS) {
			return fmt.Errorf("%q does not support this OS (%s)", name, runtime.GOOS)
		}

		if !semver.LessThan(entry.Version, it.Version) {
			fmt.Printf("%s is already up to date (%s)\n", name, entry.Version)
			continue
		}

		fmt.Printf("Updating %s (%s -> %s)...\n", name, entry.Version, it.Version)
		res, err := installer.Install(baseURL(), it)
		if err != nil {
			return fmt.Errorf("updating %s: %w", name, err)
		}

		st.Set(it.Name, it.Kind, it.Version, res.TargetPath, res.LinkPath, res.Files)
		if err := st.Save(); err != nil {
			return err
		}
		fmt.Printf("Updated %s -> %s\n", name, it.Version)
	}
	return nil
}
