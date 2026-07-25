// Package installer downloads manifest items from the tool repository and
// places them on disk according to their target_path / link_path.
package installer

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/arifcinartekin/tools/cli/internal/manifest"
)

// ExpandPath resolves a leading "~" (or "~/...") to the current user's home
// directory. Paths without a leading "~" are returned unchanged.
func ExpandPath(p string) (string, error) {
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	if p == "~" {
		return home, nil
	}
	return filepath.Join(home, strings.TrimPrefix(p, "~/")), nil
}

var httpClient = &http.Client{Timeout: 60 * time.Second}

func downloadTo(url, destPath string, executable bool) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", destPath, err)
	}

	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading %s: unexpected status %s", url, resp.Status)
	}

	mode := os.FileMode(0o644)
	if executable {
		mode = 0o755
	}

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return fmt.Errorf("creating %s: %w", destPath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("writing %s: %w", destPath, err)
	}
	return nil
}

// Result describes what was installed, for recording into local state.
type Result struct {
	TargetPath string
	LinkPath   string
	Files      []string
}

// Install downloads item's files from baseURL and places them on disk.
//
// Scripts install as a single executable file at the item's target_path.
// Apps install their full file tree under target_path (a directory) and,
// if entrypoint/link_path are set, symlink the entrypoint into link_path.
func Install(baseURL string, it manifest.Item) (Result, error) {
	targetPath, err := ExpandPath(it.TargetPath)
	if err != nil {
		return Result{}, err
	}

	switch it.Kind {
	case "script":
		if len(it.Files) != 1 {
			return Result{}, fmt.Errorf("script %q must declare exactly one file, got %d", it.Name, len(it.Files))
		}
		url := strings.TrimSuffix(baseURL, "/") + "/" + path.Join(it.BasePath, it.Files[0])
		if err := downloadTo(url, targetPath, it.Executable); err != nil {
			return Result{}, err
		}
		return Result{TargetPath: targetPath, Files: []string{targetPath}}, nil

	case "app":
		var installed []string
		for _, f := range it.Files {
			url := strings.TrimSuffix(baseURL, "/") + "/" + path.Join(it.BasePath, f)
			dest := filepath.Join(targetPath, filepath.FromSlash(f))
			executable := it.Executable && f == it.Entrypoint
			if err := downloadTo(url, dest, executable); err != nil {
				return Result{}, err
			}
			installed = append(installed, dest)
		}

		linkPath := ""
		if it.Entrypoint != "" && it.LinkPath != "" {
			linkPath, err = ExpandPath(it.LinkPath)
			if err != nil {
				return Result{}, err
			}
			entrypointPath := filepath.Join(targetPath, filepath.FromSlash(it.Entrypoint))
			if err := os.MkdirAll(filepath.Dir(linkPath), 0o755); err != nil {
				return Result{}, fmt.Errorf("creating directory for %s: %w", linkPath, err)
			}
			_ = os.Remove(linkPath) // replace any existing link/file idempotently
			if err := os.Symlink(entrypointPath, linkPath); err != nil {
				return Result{}, fmt.Errorf("linking %s -> %s: %w", linkPath, entrypointPath, err)
			}
		}

		return Result{TargetPath: targetPath, LinkPath: linkPath, Files: installed}, nil

	default:
		return Result{}, fmt.Errorf("unknown item kind %q for %q", it.Kind, it.Name)
	}
}

// Uninstall removes previously installed files and, if present, the link
// pointing at them.
func Uninstall(targetPath, linkPath string, kind string) error {
	if linkPath != "" {
		if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing link %s: %w", linkPath, err)
		}
	}
	if kind == "app" {
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("removing %s: %w", targetPath, err)
		}
		return nil
	}
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing %s: %w", targetPath, err)
	}
	return nil
}
