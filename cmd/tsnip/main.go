// Command palette for the terminal.
//
// Renders the interactive UI on stderr and prints the selected
// snippet command to stdout for shell integration.
package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aashishvinu/tsnip/data"
	"github.com/aashishvinu/tsnip/internal/app"
	"github.com/aashishvinu/tsnip/internal/search"
	"github.com/aashishvinu/tsnip/internal/storage"
	"github.com/aashishvinu/tsnip/shell"
)

func main() {
	if len(os.Args) >= 2 && os.Args[1] == "init" {
		runInit(os.Args[2:])
		return
	}

	path, err := storage.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tsnip: %v\n", err)
		os.Exit(2)
	}

	store := storage.NewJSONStore(path, data.Seed)
	doc, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "tsnip: load data: %v\n", err)
		os.Exit(2)
	}

	model := app.New(app.Config{
		Data:   doc,
		Store:  store,
		Engine: search.New(),
	})

	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithOutput(os.Stderr),
	)

	finalModel, err := program.Run()
	if err != nil {
		// Ctrl+C arrives as InterruptMsg → ErrInterrupted; treat as cancel.
		if errors.Is(err, tea.ErrInterrupted) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "tsnip: %v\n", err)
		os.Exit(2)
	}

	m, ok := finalModel.(app.Model)
	if !ok {
		os.Exit(2)
	}

	if m.Cancelled() {
		os.Exit(1)
	}

	cmd := m.SelectedCommand()
	fmt.Fprint(os.Stdout, cmd)
	if len(cmd) == 0 || cmd[len(cmd)-1] != '\n' {
		fmt.Fprintln(os.Stdout)
	}
}

func runInit(args []string) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "usage: tsnip init <zsh|bash>\n")
		os.Exit(2)
	}
	switch args[0] {
	case "zsh":
		fmt.Print(shell.Zsh)
	case "bash":
		fmt.Print(shell.Bash)
	default:
		fmt.Fprintf(os.Stderr, "tsnip: unknown shell %q (want zsh or bash)\n", args[0])
		os.Exit(2)
	}
}
