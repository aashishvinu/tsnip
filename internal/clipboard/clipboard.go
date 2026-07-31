// Package clipboard copies text to the system clipboard.
package clipboard

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Write copies text to the clipboard.
// Tries native tools first, then OSC 52 (works in many modern terminals).
func Write(text string) error {
	if err := writeNative(text); err == nil {
		return nil
	}
	return writeOSC52(os.Stderr, text)
}

func writeNative(text string) error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("clip")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	case "darwin":
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	default:
		// WSL: prefer Windows clipboard when available.
		if _, err := exec.LookPath("clip.exe"); err == nil {
			cmd := exec.Command("clip.exe")
			cmd.Stdin = strings.NewReader(strings.ReplaceAll(text, "\n", "\r\n"))
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
		for _, tool := range []string{"wl-copy", "xclip", "xsel"} {
			if _, err := exec.LookPath(tool); err != nil {
				continue
			}
			var cmd *exec.Cmd
			switch tool {
			case "wl-copy":
				cmd = exec.Command("wl-copy")
			case "xclip":
				cmd = exec.Command("xclip", "-selection", "clipboard")
			case "xsel":
				cmd = exec.Command("xsel", "--clipboard", "--input")
			}
			cmd.Stdin = strings.NewReader(text)
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
		return fmt.Errorf("no clipboard tool found")
	}
}

func writeOSC52(w io.Writer, text string) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	_, err := fmt.Fprintf(w, "\033]52;c;%s\a", encoded)
	return err
}
