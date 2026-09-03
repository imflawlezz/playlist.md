package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/imflawlezz/playlist-md/internal/engine"
	"github.com/imflawlezz/playlist-md/internal/ui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "export":
			os.Exit(runExportCLI(os.Args[2:]))
		case "core":
			args := os.Args[2:]
			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "usage: playlist-md core <command>")
				os.Exit(1)
			}
			os.Exit(engine.Forward(args))
		}
	}

	client, err := engine.NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var program *tea.Program
	notify := func(msg tea.Msg) {
		if program != nil {
			program.Send(msg)
		}
	}

	m := ui.NewModel(client, notify)
	program = tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if err := program.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runExportCLI(args []string) int {
	var output string
	var exportAll bool
	var exportLibrary bool
	var ids []string
	writeLogs := true

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--output":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "--output requires a path")
				return 1
			}
			output = args[i+1]
			i++
		case "--all":
			exportAll = true
		case "--library":
			exportLibrary = true
		case "--ids":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "--ids requires a comma-separated list")
				return 1
			}
			for _, part := range strings.Split(args[i+1], ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					ids = append(ids, part)
				}
			}
			i++
		case "--write-logs":
			writeLogs = true
		case "--no-write-logs":
			writeLogs = false
		default:
			fmt.Fprintf(os.Stderr, "unknown export flag: %s\n", args[i])
			return 1
		}
	}

	if output == "" {
		fmt.Fprintln(os.Stderr, "--output is required")
		return 1
	}
	if exportLibrary && (exportAll || len(ids) > 0) {
		fmt.Fprintln(os.Stderr, "--library cannot be combined with --all or --ids")
		return 1
	}

	client, err := engine.NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	onProgress := func(event engine.ProgressEvent) {
		switch event.Phase {
		case "fetching":
			fmt.Fprintf(os.Stderr, "Fetching [%d/%d]: %s\n", event.Index, event.Total, event.Name)
		case "library":
			fmt.Fprintln(os.Stderr, "Fetching library songs...")
		case "writing":
			fmt.Fprintln(os.Stderr, "Writing Markdown files...")
		case "cleaning":
			fmt.Fprintln(os.Stderr, "Cleaning up stale exports...")
		}
	}

	var result engine.ExportResult
	if exportLibrary {
		result, err = client.ExportLibrary(output, writeLogs, onProgress)
	} else {
		result, err = client.Export(output, ids, exportAll, writeLogs, onProgress)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
