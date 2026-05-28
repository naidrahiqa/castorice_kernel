package pages

import (
	"os"
	"os/exec"
	"path/filepath"

	"fyne.io/fyne/v2"
)

// GetLogOutputDir returns the configured directory to save pulled crash logs.
// Defaults to the user's Desktop/KernelRescue_Logs.
func GetLogOutputDir() string {
	var dir string
	if fyne.CurrentApp() != nil {
		dir = fyne.CurrentApp().Preferences().String("crash_log_dir")
	}
	if dir == "" {
		// Fallback to desktop/KernelRescue_Logs
		desktop := filepath.Join(os.Getenv("USERPROFILE"), "Desktop")
		if _, err := os.Stat(desktop); os.IsNotExist(err) {
			desktop = "."
		}
		dir = filepath.Join(desktop, "KernelRescue_Logs")
	}
	return dir
}

// SetLogOutputDir saves the path to the crash log directory in the preferences.
func SetLogOutputDir(dir string) {
	if fyne.CurrentApp() != nil {
		fyne.CurrentApp().Preferences().SetString("crash_log_dir", dir)
	}
}

// OpenFolderInExplorer opens the configured directory in Windows Explorer.
func OpenFolderInExplorer(dir string) {
	_ = os.MkdirAll(dir, 0755) // Ensure the folder exists

	var cmd *exec.Cmd
	if filepath.Separator == '\\' {
		// Windows
		cmd = exec.Command("explorer.exe", dir)
	} else {
		// Fallback for other environments (e.g. Linux / macOS)
		cmd = exec.Command("xdg-open", dir)
	}
	_ = cmd.Start()
}
