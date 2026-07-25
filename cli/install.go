package main

import (
	"fmt"
	"runtime"

	"github.com/arifcinartekin/tools/cli/internal/installer"
	"github.com/arifcinartekin/tools/cli/internal/manifest"
	"github.com/arifcinartekin/tools/cli/internal/state"
)

func runInstall(names []string) error {
	if len(names) == 0 {
		return fmt.Errorf("install requires at least one item name (see 'toolbox list')")
	}

	m, err := manifest.Fetch(baseURL())
	if err != nil {
		return err
	}

	st, err := state.Load()
	if err != nil {
		return err
	}

	for _, name := range names {
		it, ok := m.Find(name)
		if !ok {
			return fmt.Errorf("%q not found in repository (see 'toolbox list')", name)
		}
		if !it.SupportsCurrentOS(runtime.GOOS) {
			return fmt.Errorf("%q does not support this OS (%s)", name, runtime.GOOS)
		}

		fmt.Printf("Installing %s (%s)...\n", it.Name, it.Version)
		res, err := installer.Install(baseURL(), it)
		if err != nil {
			return fmt.Errorf("installing %s: %w", name, err)
		}

		st.Set(it.Name, it.Kind, it.Version, res.TargetPath, res.LinkPath, res.Files)
		if err := st.Save(); err != nil {
			return err
		}

		if res.LinkPath != "" {
			fmt.Printf("Installed %s -> %s (linked at %s)\n", it.Name, res.TargetPath, res.LinkPath)
		} else {
			fmt.Printf("Installed %s -> %s\n", it.Name, res.TargetPath)
		}
	}
	return nil
}
