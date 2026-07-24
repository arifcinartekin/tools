// Package manifest fetches and parses the central manifest.json published
// alongside the tool repository (built by CI from each item's meta.json).
package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Item describes a single installable script or app.
type Item struct {
	Name        string   `json:"name"`
	Kind        string   `json:"kind"` // "script" or "app"
	Description string   `json:"description"`
	Version     string   `json:"version"`
	OS          []string `json:"os"`
	TargetPath  string   `json:"target_path"`
	Entrypoint  string   `json:"entrypoint"`
	LinkPath    string   `json:"link_path,omitempty"`
	Executable  bool     `json:"executable"`
	BasePath    string   `json:"base_path"`
	Files       []string `json:"files"`
}

// Manifest is the top-level document produced by build-manifest.sh.
type Manifest struct {
	GeneratedAt string `json:"generated_at"`
	Items       []Item `json:"items"`
}

// Fetch downloads and parses manifest.json from baseURL (e.g.
// https://pkg.arifcinartekin.me).
func Fetch(baseURL string) (*Manifest, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(baseURL + "/manifest.json")
	if err != nil {
		return nil, fmt.Errorf("fetching manifest: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching manifest: unexpected status %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}
	return &m, nil
}

// Find returns the item with the given name, if present.
func (m *Manifest) Find(name string) (Item, bool) {
	for _, it := range m.Items {
		if it.Name == name {
			return it, true
		}
	}
	return Item{}, false
}

// SupportsCurrentOS reports whether the item declares support for goos
// ("linux" or "darwin"), or declares no restriction at all.
func (it Item) SupportsCurrentOS(goos string) bool {
	if len(it.OS) == 0 {
		return true
	}
	for _, o := range it.OS {
		if o == goos || o == "any" {
			return true
		}
	}
	return false
}
