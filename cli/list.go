package main

import (
	"fmt"
	"runtime"
	"sort"

	"github.com/arifcinartekin/tools/cli/internal/manifest"
	"github.com/arifcinartekin/tools/cli/internal/state"
)

func runList() error {
	m, err := manifest.Fetch(baseURL())
	if err != nil {
		return err
	}

	st, err := state.Load()
	if err != nil {
		return err
	}

	items := append([]manifest.Item(nil), m.Items...)
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })

	fmt.Printf("%-24s %-6s %-10s %-10s %s\n", "NAME", "KIND", "VERSION", "INSTALLED", "DESCRIPTION")
	for _, it := range items {
		installedVersion := "-"
		if entry, ok := st.Items[it.Name]; ok {
			installedVersion = entry.Version
		}
		supported := ""
		if !it.SupportsCurrentOS(runtime.GOOS) {
			supported = " (unsupported on this OS)"
		}
		fmt.Printf("%-24s %-6s %-10s %-10s %s%s\n", it.Name, it.Kind, it.Version, installedVersion, it.Description, supported)
	}
	return nil
}
