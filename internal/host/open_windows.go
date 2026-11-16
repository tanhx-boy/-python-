//go:build windows

package host

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

// Open launches url in a standalone browser app window when Edge or Chrome is
// available, falling back to the default browser.
func Open(url string) error {
	var candidates []string
	for _, env := range []string{"ProgramFiles(x86)", "ProgramFiles", "LOCALAPPDATA"} {
		base := os.Getenv(env)
		if base == "" {
			continue
		}
		candidates = append(candidates,
			filepath.Join(base, "Microsoft", "Edge", "Application", "msedge.exe"),
			filepath.Join(base, "Google", "Chrome", "Application", "chrome.exe"),
		)
	}

	for _, exe := range candidates {
		info, err := os.Stat(exe)
		if err != nil || info.IsDir() {
			continue
		}
		cmd := exec.Command(exe,
			"--app="+url,
			"--start-maximized",
			"--window-size=1280,860",
			"--no-first-run",
			"--no-default-browser-check",
		)
		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

// Notice displays a blocking native message box. The GUI build has no console,
// so startup failures must be surfaced this way.
func Notice(title, text string) {
	const iconWarning = 0x00000030
	titlePtr, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return
	}
	textPtr, err := syscall.UTF16PtrFromString(text)
	if err != nil {
		return
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	_, _, _ = proc.Call(0,
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		iconWarning,
	)
}
