// Command toolbox installs and updates scripts and apps published in the
// arifcinartekin/pkg tool repository, and manages the local record of what
// is currently installed.
package main

import (
	"fmt"
	"os"
)

// defaultBaseURL is the GitHub Pages site the manifest and item files are
// published from. Override with the TOOLBOX_BASE_URL environment variable
// (useful for local testing against a repo checkout served over HTTP).
const defaultBaseURL = "https://pkg.arifcinartekin.me"

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func baseURL() string {
	if v := os.Getenv("TOOLBOX_BASE_URL"); v != "" {
		return v
	}
	return defaultBaseURL
}

func usage() {
	fmt.Fprint(os.Stderr, `toolbox - personal script/app package manager

Usage:
  toolbox list                    List everything available in the repository
  toolbox install <name>...       Install one or more items
  toolbox update <name>...        Update one or more installed items
  toolbox update --all            Update every installed item
  toolbox uninstall <name>...     Uninstall one or more items
  toolbox version                 Print the toolbox version

Environment:
  TOOLBOX_BASE_URL   Override the repository base URL (default: https://pkg.arifcinartekin.me)
`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "list", "ls":
		err = runList()
	case "install", "add":
		err = runInstall(os.Args[2:])
	case "update", "upgrade":
		err = runUpdate(os.Args[2:])
	case "uninstall", "remove", "rm":
		err = runUninstall(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Println("toolbox " + version)
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "toolbox: unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "toolbox: %v\n", err)
		os.Exit(1)
	}
}
