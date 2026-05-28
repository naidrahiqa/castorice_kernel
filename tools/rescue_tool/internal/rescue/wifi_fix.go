package rescue

import (
	"fmt"
	"strings"

	"github.com/naidrahiqa/epitaph_rescue/internal/adb"
	"github.com/naidrahiqa/epitaph_rescue/internal/logger"
)

// WiFiFixStep represents a step in the WiFi fix flow
type WiFiFixStep int

const (
	WiFiStepDiagnose     WiFiFixStep = iota // Diagnose the situation
	WiFiStepExplain                          // Explain why WiFi is broken
	WiFiStepGuideFlash                       // Guide: flash stock boot via KernelFlasher
	WiFiStepVerify                           // Verify WiFi is working
	WiFiStepReflash                          // Guide: reflash Epitaph ZIP
	WiFiStepDone                             // Done
)

func (s WiFiFixStep) Title() string {
	switch s {
	case WiFiStepDiagnose:
		return "Diagnosa Masalah WiFi"
	case WiFiStepExplain:
		return "Kenapa WiFi Mati?"
	case WiFiStepGuideFlash:
		return "Flash Stock Boot via KernelFlasher"
	case WiFiStepVerify:
		return "Verifikasi WiFi"
	case WiFiStepReflash:
		return "Flash Ulang Custom Kernel"
	case WiFiStepDone:
		return "Selesai!"
	default:
		return ""
	}
}

// WiFiDiagnosis holds the results of WiFi diagnosis
type WiFiDiagnosis struct {
	WiFiInterface  string // wlan0 status
	WiFiModules    string // loaded WiFi modules
	VendorDLKM     string // vendor_dlkm mount status
	IsWiFiBroken   bool
	IsFastbootFlash bool // was it flashed via fastboot?
	RootCause      string
}

// WiFiFixer handles WiFi fix logic
type WiFiFixer struct {
	deviceMgr *adb.DeviceManager
	diagnosis *WiFiDiagnosis
	onUpdate  func()
}

// NewWiFiFixer creates a new WiFi fixer
func NewWiFiFixer(dm *adb.DeviceManager) *WiFiFixer {
	return &WiFiFixer{
		deviceMgr: dm,
		diagnosis: &WiFiDiagnosis{},
	}
}

func (wf *WiFiFixer) SetOnUpdate(fn func()) { wf.onUpdate = fn }
func (wf *WiFiFixer) Diagnosis() *WiFiDiagnosis { return wf.diagnosis }

// DiagnoseWiFi checks WiFi status on the connected device
func (wf *WiFiFixer) DiagnoseWiFi() error {
	info := wf.deviceMgr.Detect()
	if info.Mode != adb.ModeAndroid {
		return fmt.Errorf("device tidak terhubung di Android mode")
	}

	adbClient := wf.deviceMgr.GetADBClient()

	// Check wlan0 interface
	iface, _ := adbClient.Shell("ip link show wlan0 2>/dev/null || echo 'NOT_FOUND'")
	wf.diagnosis.WiFiInterface = strings.TrimSpace(iface)

	// Check loaded WiFi modules
	modules, _ := adbClient.Shell("lsmod 2>/dev/null | grep -i 'wlan\\|wifi\\|cfg80211\\|mac80211' || echo 'NONE'")
	wf.diagnosis.WiFiModules = strings.TrimSpace(modules)

	// Check vendor_dlkm
	dlkm, _ := adbClient.Shell("mount | grep vendor_dlkm 2>/dev/null || echo 'NOT_MOUNTED'")
	wf.diagnosis.VendorDLKM = strings.TrimSpace(dlkm)

	// Determine if WiFi is broken
	if strings.Contains(wf.diagnosis.WiFiInterface, "NOT_FOUND") ||
		strings.Contains(wf.diagnosis.WiFiModules, "NONE") {
		wf.diagnosis.IsWiFiBroken = true
		wf.diagnosis.RootCause = "WiFi interface (wlan0) tidak ditemukan. " +
			"Ini terjadi karena modul WiFi di /vendor_dlkm tidak cocok dengan kernel yang sedang berjalan."
	}

	// Check if flashed via fastboot (kernel version hint)
	kernel, _ := adbClient.Shell("uname -r 2>/dev/null")
	if strings.Contains(strings.ToLower(kernel), "stock") || strings.Contains(strings.ToLower(kernel), "generic") {
		wf.diagnosis.IsFastbootFlash = true
	}

	logger.Info("WiFi diagnosis: broken=%v, fastbootFlash=%v", wf.diagnosis.IsWiFiBroken, wf.diagnosis.IsFastbootFlash)
	return nil
}

// GetExplanation returns a human-readable explanation of why WiFi is broken
func (wf *WiFiFixer) GetExplanation() string {
	return `🔍 Kenapa WiFi Mati?

Saat kamu flash kernel via Fastboot, yang terjadi adalah:

1. Fastboot hanya replace partition "boot" saja
2. Partition "vendor_dlkm" (berisi modul WiFi) TIDAK di-update
3. Modul WiFi di vendor_dlkm punya "vermagic" yang harus match persis dengan kernel
4. Karena kernel baru punya vermagic berbeda → modul WiFi DITOLAK → WiFi mati

📌 Solusi: Flash kernel menggunakan KernelFlasher (bukan Fastboot)
   KernelFlasher akan otomatis update modul di vendor_dlkm supaya cocok.

Langkah-langkah fix ada di halaman selanjutnya →`
}

// GetFlashGuide returns step-by-step instructions to fix WiFi
func (wf *WiFiFixer) GetFlashGuide() string {
	return `📋 Langkah Fix WiFi:

Step 1: Download stock boot.img untuk ROM kamu
   → Bisa dari group Telegram atau website resmi

Step 2: Install KernelFlasher dari GitHub
   → Buka KernelFlasher di HP
   → Perlu root (KSU atau Magisk)

Step 3: Flash stock boot.img via KernelFlasher
   → JANGAN pakai Fastboot!
   → KernelFlasher → pilih boot.img → Flash
   → Reboot

Step 4: Cek WiFi nyala
   → Buka Settings → WiFi → pastikan bisa scan network

Step 5: Flash Custom Kernel ZIP via KernelFlasher
   → Download Custom Kernel ZIP terbaru
   → KernelFlasher → AnyKernel3 → pilih ZIP → Flash
   → Reboot

✅ WiFi seharusnya normal setelah ini karena KernelFlasher 
   akan update vendor_dlkm secara otomatis.`
}
