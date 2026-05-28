package main

import (
	"image/color"
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/naidrahiqa/epitaph_rescue/internal/adb"
	"github.com/naidrahiqa/epitaph_rescue/internal/logger"
	"github.com/naidrahiqa/epitaph_rescue/internal/tools"
	"github.com/naidrahiqa/epitaph_rescue/ui/pages"
	"github.com/naidrahiqa/epitaph_rescue/ui/theme"
)

const (
	AppVersion = "0.1.0"
	AppTitle   = "Kernel Rescue Tool"
)

func main() {
	// Initialize logging
	appLog, err := logger.Init()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLog.Close()

	logger.Info("========================================")
	logger.Info("Epitaph Rescue Tool %s starting...", AppVersion)
	logger.Info("========================================")

	// Ensure platform-tools are available
	toolsManager := tools.NewManager()
	adbPath, fastbootPath := toolsManager.GetPaths()

	logger.Info("ADB path: %s", adbPath)
	logger.Info("Fastboot path: %s", fastbootPath)

	// Initialize ADB client
	adbClient := adb.NewClient(adbPath)
	fastbootClient := adb.NewFastbootClient(fastbootPath)
	deviceManager := adb.NewDeviceManager(adbClient, fastbootClient)

	// Create Fyne app
	fyneApp := app.NewWithID("com.naidrahiqa.epitaph_rescue")
	fyneApp.Settings().SetTheme(theme.NewEpitaphTheme())

	window := fyneApp.NewWindow(AppTitle + " v" + AppVersion)
	window.Resize(fyne.NewSize(720, 580))
	window.SetFixedSize(false)
	window.CenterOnScreen()

	// Create pages
	homePage := pages.NewHomePage(deviceManager, toolsManager, window)
	rescuePage := pages.NewRescuePage(deviceManager, window)
	logPage := pages.NewLogViewerPage(deviceManager, window)
	wifiPage := pages.NewWiFiFixPage(deviceManager, window)
	flasherPage := pages.NewFlasherPage(deviceManager, window)
	toolboxPage := pages.NewToolboxPage(deviceManager, window)

	// Navigation tabs
	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Home", theme.HomeIcon(), homePage.Content()),
		container.NewTabItemWithIcon("Rescue", theme.RescueIcon(), rescuePage.Content()),
		container.NewTabItemWithIcon("Log", theme.LogIcon(), logPage.Content()),
		container.NewTabItemWithIcon("Diagnose", theme.SearchIcon(), wifiPage.Content()),
		container.NewTabItemWithIcon("Flasher", theme.ValidateIcon(), flasherPage.Content()),
		container.NewTabItemWithIcon("Toolbox", theme.SettingsIcon(), toolboxPage.Content()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// Setup tab navigation callback from Home Dashboard
	homePage.SetOnNavigate(func(index int) {
		tabs.SelectTabIndex(index)
	})
	rescuePage.SetOnNavigate(func(index int) {
		tabs.SelectTabIndex(index)
	})
	wifiPage.SetOnNavigate(func(index int) {
		tabs.SelectTabIndex(index)
	})

	// Status bar at bottom
	statusLabel := widget.NewLabel("Epitaph Rescue Tool v" + AppVersion + " — Ready")
	statusLabel.TextStyle = fyne.TextStyle{Bold: true}
	statusBg := canvas.NewRectangle(color.RGBA{R: 0x12, G: 0x16, B: 0x20, A: 0xff})
	statusBorder := canvas.NewRectangle(color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff})
	statusBorder.SetMinSize(fyne.NewSize(0, 2))
	statusBar := container.NewBorder(nil, statusBorder, nil, nil,
		container.NewPadded(statusLabel),
	)
	statusBarStack := container.New(layout.NewMaxLayout(), statusBg, statusBar)

	mainContent := container.NewBorder(
		nil,           // top
		statusBarStack, // bottom
		nil,           // left
		nil,           // right
		tabs,          // center
	)

	window.SetContent(mainContent)

	// Start device polling
	homePage.StartPolling()

	// Cleanup on close
	window.SetOnClosed(func() {
		homePage.StopPolling()
		logger.Info("Application closed")
	})

	// Ensure cleanup runs
	window.SetCloseIntercept(func() {
		homePage.StopPolling()
		logger.Info("Application closing...")
		window.Close()
	})

	logger.Info("Application ready, showing window")
	window.ShowAndRun()
}

// getAppDataDir returns the path to the app's data directory
func getAppDataDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = "."
	}
	return filepath.Join(appData, "KernelRescueTool")
}
