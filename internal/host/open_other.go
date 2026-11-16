//go:build !windows

package host

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Open launches url in the platform's default browser.
func Open(url string) error {
	launchers := []string{"xdg-open", "gio", "sensible-browser", "x-www-browser"}
	if runtime.GOOS == "darwin" {
		launchers = []string{"open"}
	}
	for _, name := range launchers {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}
		args := []string{url}
		if name == "gio" {
			args = []string{"open", url}
		}
		cmd := exec.Command(path, args...)
		if err := cmd.Start(); err == nil {
			go func() { _ = cmd.Wait() }()
			return nil
		}
	}
	return fmt.Errorf("no supported browser launcher found")
}

// Notice writes a message to stderr on platforms with a console.
func Notice(title, text string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", title, text)
}
