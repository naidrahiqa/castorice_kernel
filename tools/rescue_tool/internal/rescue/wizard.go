package rescue

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/naidrahiqa/epitaph_rescue/internal/adb"
	"github.com/naidrahiqa/epitaph_rescue/internal/logger"
	"github.com/naidrahiqa/epitaph_rescue/internal/parser"
)

// WizardStep represents a step in the rescue flow
type WizardStep int

const (
	StepDetect     WizardStep = iota // Step 1: Detect device situation
	StepFlashBoot                     // Step 2: Flash stock boot.img
	StepWaitReboot                    // Step 3: Wait for reboot to Android
	StepPullLog                       // Step 4: Pull crash log
	StepAnalyze                       // Step 5: Parse & analyze log
	StepComplete                      // Done
)

func (s WizardStep) String() string {
	switch s {
	case StepDetect:
		return "Detect Device"
	case StepFlashBoot:
		return "Flash Stock Boot"
	case StepWaitReboot:
		return "Wait for Reboot"
	case StepPullLog:
		return "Pull Crash Log"
	case StepAnalyze:
		return "Analyze Log"
	case StepComplete:
		return "Complete"
	default:
		return "Unknown"
	}
}

func (s WizardStep) Description() string {
	switch s {
	case StepDetect:
		return "Mendeteksi status device — Fastboot atau Android?"
	case StepFlashBoot:
		return "Flash stock boot.img untuk mengembalikan HP ke kondisi booting normal"
	case StepWaitReboot:
		return "Menunggu device reboot ke Android... (timeout 3 menit)"
	case StepPullLog:
		return "Menarik crash log dari PStore/RAMoops"
	case StepAnalyze:
		return "Menganalisis crash log untuk menemukan penyebab bootloop"
	case StepComplete:
		return "Rescue selesai! Lihat hasil analisis di bawah."
	default:
		return ""
	}
}

// StepStatus tracks the status of each wizard step
type StepStatus int

const (
	StatusPending  StepStatus = iota
	StatusRunning
	StatusSuccess
	StatusFailed
	StatusSkipped
)

// WizardState holds the full state of the rescue wizard
type WizardState struct {
	mu             sync.RWMutex
	CurrentStep    WizardStep
	StepStatuses   map[WizardStep]StepStatus
	StepMessages   map[WizardStep]string
	LogFilePath    string
	LogContent     string
	AnalysisResult *parser.AnalysisResult
	Error          error
}

// NewWizardState initializes a fresh wizard state
func NewWizardState() *WizardState {
	ws := &WizardState{
		CurrentStep:  StepDetect,
		StepStatuses: make(map[WizardStep]StepStatus),
		StepMessages: make(map[WizardStep]string),
	}
	for step := StepDetect; step <= StepComplete; step++ {
		ws.StepStatuses[step] = StatusPending
	}
	return ws
}

func (ws *WizardState) SetStep(step WizardStep, status StepStatus, msg string) {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	ws.CurrentStep = step
	ws.StepStatuses[step] = status
	ws.StepMessages[step] = msg
}

func (ws *WizardState) GetStep() (WizardStep, StepStatus, string) {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	step := ws.CurrentStep
	return step, ws.StepStatuses[step], ws.StepMessages[step]
}

func (ws *WizardState) GetAnalysisResult() *parser.AnalysisResult {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.AnalysisResult
}

func (ws *WizardState) GetLogFilePath() string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.LogFilePath
}

func (ws *WizardState) GetLogContent() string {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.LogContent
}

// Wizard orchestrates the rescue flow
type Wizard struct {
	deviceMgr    *adb.DeviceManager
	parser       *parser.Parser
	state        *WizardState
	onUpdate     func() // Callback to refresh UI
	LogOutputDir string
}

// NewWizard creates a new rescue wizard
func NewWizard(dm *adb.DeviceManager) *Wizard {
	return &Wizard{
		deviceMgr: dm,
		parser:    parser.NewParser(),
		state:     NewWizardState(),
	}
}

func (w *Wizard) SetLogOutputDir(dir string) {
	w.LogOutputDir = dir
}

func (w *Wizard) GetLogOutputDir() string {
	if w.LogOutputDir == "" {
		desktop := filepath.Join(os.Getenv("USERPROFILE"), "Desktop")
		if _, err := os.Stat(desktop); os.IsNotExist(err) {
			desktop = "."
		}
		return filepath.Join(desktop, "EpitaphRescue_Logs")
	}
	return w.LogOutputDir
}

func (w *Wizard) State() *WizardState { return w.state }

func (w *Wizard) SetOnUpdate(fn func()) { w.onUpdate = fn }

func (w *Wizard) notify() {
	if w.onUpdate != nil {
		w.onUpdate()
	}
}

// Reset resets the wizard to initial state
func (w *Wizard) Reset() {
	w.state = NewWizardState()
	w.notify()
}

// RunStep1Detect detects device status
func (w *Wizard) RunStep1Detect() {
	w.state.SetStep(StepDetect, StatusRunning, "Mendeteksi device...")
	w.notify()

	info := w.deviceMgr.Detect()

	switch info.Mode {
	case adb.ModeFastboot:
		w.state.SetStep(StepDetect, StatusSuccess,
			fmt.Sprintf("Device terdeteksi di Fastboot Mode (serial: %s)", info.Serial))
		logger.Info("Wizard Step 1: Device in Fastboot mode")
	case adb.ModeAndroid:
		w.state.SetStep(StepDetect, StatusSuccess,
			fmt.Sprintf("Device terdeteksi di Android Mode (%s) — langsung ke Pull Log", info.Model))
		// Skip flash & reboot steps
		w.state.StepStatuses[StepFlashBoot] = StatusSkipped
		w.state.StepMessages[StepFlashBoot] = "Dilewati — device sudah di Android"
		w.state.StepStatuses[StepWaitReboot] = StatusSkipped
		w.state.StepMessages[StepWaitReboot] = "Dilewati — device sudah di Android"
		logger.Info("Wizard Step 1: Device in Android mode, skipping flash steps")
	default:
		w.state.SetStep(StepDetect, StatusFailed,
			"Device tidak terdeteksi! Pastikan USB terhubung dan driver terinstall.")
		logger.Warn("Wizard Step 1: No device detected")
	}
	w.notify()
}

// RunStep2Flash flashes stock boot.img via Fastboot
func (w *Wizard) RunStep2Flash(bootImgPath string) {
	w.state.SetStep(StepFlashBoot, StatusRunning, "Memvalidasi boot.img...")
	w.notify()

	// Validate file exists
	if _, err := os.Stat(bootImgPath); os.IsNotExist(err) {
		w.state.SetStep(StepFlashBoot, StatusFailed, "File boot.img tidak ditemukan!")
		w.notify()
		return
	}

	// Validate magic bytes (Android boot image starts with "ANDROID!")
	f, err := os.Open(bootImgPath)
	if err != nil {
		w.state.SetStep(StepFlashBoot, StatusFailed, fmt.Sprintf("Gagal buka file: %v", err))
		w.notify()
		return
	}
	magic := make([]byte, 8)
	n, _ := f.Read(magic)
	f.Close()

	if n < 8 || string(magic) != "ANDROID!" {
		w.state.SetStep(StepFlashBoot, StatusFailed,
			"File bukan boot.img yang valid! (Magic bytes tidak cocok — harus 'ANDROID!')")
		w.notify()
		return
	}

	w.state.SetStep(StepFlashBoot, StatusRunning,
		fmt.Sprintf("Flashing %s ke partition boot...", filepath.Base(bootImgPath)))
	w.notify()

	fb := w.deviceMgr.GetFastbootClient()
	output, err := fb.Flash("boot", bootImgPath)
	if err != nil {
		w.state.SetStep(StepFlashBoot, StatusFailed,
			fmt.Sprintf("Flash gagal: %v\nOutput: %s", err, output))
		w.notify()
		return
	}

	w.state.SetStep(StepFlashBoot, StatusSuccess,
		fmt.Sprintf("Flash berhasil! Output: %s", output))
	logger.Info("Wizard Step 2: Flash complete — %s", output)
	w.notify()

	// Auto-reboot after flash
	w.state.SetStep(StepFlashBoot, StatusSuccess, "Flash selesai! Rebooting device...")
	w.notify()
	if err := fb.Reboot(); err != nil {
		logger.Warn("Reboot after flash failed: %v", err)
	}
}

// RunStep3WaitReboot polls for ADB connection after reboot
func (w *Wizard) RunStep3WaitReboot() {
	w.state.SetStep(StepWaitReboot, StatusRunning, "Menunggu device reboot ke Android...")
	w.notify()

	timeout := 3 * time.Minute
	start := time.Now()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(start)
			remaining := timeout - elapsed
			if remaining <= 0 {
				w.state.SetStep(StepWaitReboot, StatusFailed,
					"Timeout! Device tidak booting ke Android dalam 3 menit. Kemungkinan bootloop.")
				w.notify()
				return
			}

			info := w.deviceMgr.Detect()
			if info.Mode == adb.ModeAndroid {
				w.state.SetStep(StepWaitReboot, StatusSuccess,
					fmt.Sprintf("Device sudah di Android! (%s)", info.Model))
				logger.Info("Wizard Step 3: Device rebooted to Android")
				w.notify()
				return
			}

			w.state.SetStep(StepWaitReboot, StatusRunning,
				fmt.Sprintf("Menunggu reboot... (sisa %d detik)", int(remaining.Seconds())))
			w.notify()
		}
	}
}

// RunStep4PullLog pulls crash log from PStore
func (w *Wizard) RunStep4PullLog() {
	w.state.SetStep(StepPullLog, StatusRunning, "Menarik crash log dari PStore...")
	w.notify()

	adbClient := w.deviceMgr.GetADBClient()

	// Try console-ramoops first
	pstorePaths := []string{
		"/sys/fs/pstore/console-ramoops-0",
		"/sys/fs/pstore/dmesg-ramoops-0",
		"/sys/fs/pstore/console-ramoops",
		"/proc/last_kmsg",
	}

	var logContent string
	var foundPath string

	for _, remotePath := range pstorePaths {
		w.state.SetStep(StepPullLog, StatusRunning,
			fmt.Sprintf("Mencoba: %s ...", remotePath))
		w.notify()

		content, err := adbClient.Shell(fmt.Sprintf("cat %s 2>/dev/null", remotePath))
		if err == nil && strings.TrimSpace(content) != "" && len(content) > 100 {
			logContent = content
			foundPath = remotePath
			break
		}
	}

	if logContent == "" {
		w.state.SetStep(StepPullLog, StatusFailed,
			"Tidak ada crash log ditemukan di PStore. Kemungkinan PStore/RAMoops tidak aktif di kernel ini.")
		w.notify()
		return
	}

	// Save to configured directory
	outputDir := w.GetLogOutputDir()
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Error("Failed to create log directory: %v", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("last_kmsg_%s.txt", timestamp)
	localPath := filepath.Join(outputDir, fileName)

	if err := os.WriteFile(localPath, []byte(logContent), 0644); err != nil {
		logger.Error("Failed to save log file: %v", err)
		// Still continue — we have the content in memory
	}

	w.state.mu.Lock()
	w.state.LogFilePath = localPath
	w.state.LogContent = logContent
	w.state.mu.Unlock()

	w.state.SetStep(StepPullLog, StatusSuccess,
		fmt.Sprintf("Log berhasil ditarik dari %s\nDisimpan ke: %s\n(%d baris)",
			foundPath, localPath, len(strings.Split(logContent, "\n"))))
	logger.Info("Wizard Step 4: Log pulled from %s (%d bytes)", foundPath, len(logContent))
	w.notify()
}

// RunStep5Analyze parses and analyzes the pulled log
func (w *Wizard) RunStep5Analyze() {
	w.state.SetStep(StepAnalyze, StatusRunning, "Menganalisis crash log...")
	w.notify()

	w.state.mu.RLock()
	content := w.state.LogContent
	w.state.mu.RUnlock()

	if content == "" {
		w.state.SetStep(StepAnalyze, StatusFailed,
			"Tidak ada log content untuk dianalisis.")
		w.notify()
		return
	}

	result := w.parser.ParseLines(strings.Split(content, "\n"))

	w.state.mu.Lock()
	w.state.AnalysisResult = result
	w.state.mu.Unlock()

	summary := fmt.Sprintf("Analisis selesai!\n• %d baris diproses\n• %d pattern match ditemukan\n• %d CRITICAL, %d WARNING, %d INFO",
		result.TotalLines,
		len(result.Matches),
		result.SeverityCounts[parser.CRITICAL],
		result.SeverityCounts[parser.WARNING],
		result.SeverityCounts[parser.INFO],
	)

	w.state.SetStep(StepAnalyze, StatusSuccess, summary)
	w.state.CurrentStep = StepComplete
	w.state.StepStatuses[StepComplete] = StatusSuccess
	w.state.StepMessages[StepComplete] = "Rescue flow selesai!"
	logger.Info("Wizard Step 5: Analysis complete")
	w.notify()
}
