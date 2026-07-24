// Package state manages the local record of what toolbox has installed,
// stored at ~/.toolbox/installed.json.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry records what toolbox installed locally for one item.
type Entry struct {
	Kind        string   `json:"kind"`
	Version     string   `json:"version"`
	TargetPath  string   `json:"target_path"`
	LinkPath    string   `json:"link_path,omitempty"`
	Files       []string `json:"files"`
	InstalledAt string   `json:"installed_at"`
}

// State is the on-disk record of installed items, keyed by item name.
type State struct {
	Items map[string]Entry `json:"items"`

	path string
}

// Dir returns the toolbox state directory (~/.toolbox), creating it if needed.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	dir := filepath.Join(home, ".toolbox")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating state directory: %w", err)
	}
	return dir, nil
}

// Load reads installed.json, returning an empty State if it does not exist yet.
func Load() (*State, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "installed.json")

	s := &State{Items: map[string]Entry{}, path: path}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}
	s.path = path
	return s, nil
}

// Save writes the state back to installed.json.
func (s *State) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding state file: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	return nil
}

// Set records or replaces the entry for name.
func (s *State) Set(name, kind, version, targetPath, linkPath string, files []string) {
	s.Items[name] = Entry{
		Kind:        kind,
		Version:     version,
		TargetPath:  targetPath,
		LinkPath:    linkPath,
		Files:       files,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// Remove deletes the entry for name, if present.
func (s *State) Remove(name string) {
	delete(s.Items, name)
}
