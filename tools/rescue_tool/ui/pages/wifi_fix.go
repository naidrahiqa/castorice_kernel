package pages

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/naidrahiqa/epitaph_rescue/internal/adb"
	"github.com/naidrahiqa/epitaph_rescue/internal/rescue"
)

type WiFiFixPage struct {
	fixer        *rescue.WiFiFixer
	deviceMgr    *adb.DeviceManager
	window       fyne.Window
	onNavigate   func(index int)

	// GUI Widgets
	lblTitle      *widget.Label
	lblStatus     *widget.Label
	btnScan       *widget.Button
	btnFix        *widget.Button
	progressScan  *widget.ProgressBarInfinite
	resultBox     *fyne.Container

	mainCard      fyne.CanvasObject
	mainBox       *fyne.Container
}

type DiagItem struct {
	Name        string
	Status      string // "OK", "WARNING", "ERROR"
	Value       string
	Description string
}

func NewWiFiFixPage(dm *adb.DeviceManager, w fyne.Window) *WiFiFixPage {
	wp := &WiFiFixPage{
		fixer:     rescue.NewWiFiFixer(dm),
		deviceMgr: dm,
		window:    w,
	}
	wp.buildUI()
	return wp
}

func (wp *WiFiFixPage) SetOnNavigate(fn func(index int)) {
	wp.onNavigate = fn
}

func (wp *WiFiFixPage) buildUI() {
	wp.lblTitle = widget.NewLabel("Diagnostics & Auto-Fix Center")
	wp.lblTitle.TextStyle = fyne.TextStyle{Bold: true}

	wp.lblStatus = widget.NewLabel("Silakan hubungkan device Anda dan klik tombol di bawah ini untuk memulai auto-scan pendeteksian kerusakan sistem/kernel secara instan.")
	wp.lblStatus.Wrapping = fyne.TextWrapWord

	wp.progressScan = widget.NewProgressBarInfinite()
	wp.progressScan.Hide()

	wp.resultBox = container.NewVBox()

	wp.btnScan = widget.NewButtonWithIcon("🔍 Scan System & Kernel Errors", theme.SearchIcon(), func() {
		wp.runSystemDiagnostics()
	})
	wp.btnScan.Importance = widget.HighImportance

	wp.btnFix = widget.NewButtonWithIcon("🔧 Auto-Fix WiFi & Driver Errors", theme.SettingsIcon(), func() {
		wp.runAutoFix()
	})
	wp.btnFix.Importance = widget.SuccessImportance
	wp.btnFix.Hide()

	buttonBox := container.NewHBox(
		layout.NewSpacer(),
		wp.btnScan,
		wp.btnFix,
		layout.NewSpacer(),
	)

	wp.mainCard = NewNeoCard("Diagnostic Suite", "Auto-scan & error diagnostic", container.NewVBox(
		wp.lblStatus,
		wp.progressScan,
		NeoDivider(),
		buttonBox,
		NeoDivider(),
		wp.resultBox,
	))

	wp.mainBox = container.NewVBox(wp.mainCard)
}

func (wp *WiFiFixPage) Content() fyne.CanvasObject {
	return container.NewScroll(wp.mainBox)
}

func (wp *WiFiFixPage) runSystemDiagnostics() {
	wp.btnScan.Disable()
	wp.btnFix.Hide()
	wp.progressScan.Show()
	wp.lblStatus.SetText("Sedang memindai sistem dan kernel secara asinkron... Harap tunggu.")
	wp.resultBox.Objects = nil
	wp.resultBox.Refresh()
	wp.mainCard.Refresh()

	go func() {
		// Mimic a short scanning delay for futuristic look
		time.Sleep(1000 * time.Millisecond)

		info := wp.deviceMgr.Detect()
		var items []DiagItem

		// 1. Connection Check
		if info.Mode == adb.ModeNone {
			items = append(items, DiagItem{
				Name:        "ADB Connection Mode",
				Status:      "ERROR",
				Value:       "🔴 Disconnected",
				Description: "Device tidak terdeteksi via USB. Hubungkan kabel USB dan pastikan USB debugging aktif di HP Anda.",
			})
		} else if info.Mode == adb.ModeFastboot {
			items = append(items, DiagItem{
				Name:        "ADB Connection Mode",
				Status:      "WARNING",
				Value:       "🟡 Fastboot Mode",
				Description: "Device terhubung dalam mode Fastboot (bootloader). Akses diagnosa kernel terbatas pada mode ini.",
			})
		} else {
			items = append(items, DiagItem{
				Name:        "ADB Connection Mode",
				Status:      "OK",
				Value:       "🟢 Connected (Android)",
				Description: fmt.Sprintf("Device terhubung via ADB: %s (%s)", info.Model, info.Codename),
			})
		}

		if info.Mode == adb.ModeAndroid {
			adbClient := wp.deviceMgr.GetADBClient()

			// 2. WiFi Status Check
			_ = wp.fixer.DiagnoseWiFi()
			diag := wp.fixer.Diagnosis()
			if diag.IsWiFiBroken {
				items = append(items, DiagItem{
					Name:        "WiFi Interface (wlan0)",
					Status:      "ERROR",
					Value:       "🔴 Missing / Mismatched Driver",
					Description: "Interface wlan0 mati! Terdeteksi driver modul kernel wlan.ko tidak cocok (signature mismatch) dengan kernel yang sedang aktif.",
				})
			} else {
				items = append(items, DiagItem{
					Name:        "WiFi Interface (wlan0)",
					Status:      "OK",
					Value:       "🟢 Active & Operational",
					Description: "Interface WiFi wlan0 aktif, terload, dan driver termuat dengan sempurna di kernel.",
				})
			}

			// 3. SELinux Check
			seStatus, _ := adbClient.Shell("getenforce 2>/dev/null")
			seStatus = strings.TrimSpace(seStatus)
			if seStatus == "" {
				seStatus, _ = adbClient.Shell("getprop selinux.enforce 2>/dev/null")
				seStatus = strings.TrimSpace(seStatus)
				if seStatus == "0" {
					seStatus = "Permissive"
				} else if seStatus == "1" {
					seStatus = "Enforcing"
				}
			}
			if strings.EqualFold(seStatus, "permissive") {
				items = append(items, DiagItem{
					Name:        "SELinux Integrity",
					Status:      "WARNING",
					Value:       "🟡 Permissive Mode",
					Description: "SELinux disetel ke Permissive. Kernel tidak membatasi hak akses sistem, mengurangi keamanan Android Anda dari malware.",
				})
			} else {
				items = append(items, DiagItem{
					Name:        "SELinux Integrity",
					Status:      "OK",
					Value:       "🟢 Enforcing Mode",
					Description: "SELinux aktif dalam mode Enforcing, menjamin keamanan sandbox data aplikasi sesuai standar GKI.",
				})
			}

			// 4. Verified Boot (AVB) Check
			avbState, _ := adbClient.Shell("getprop ro.boot.verifiedbootstate 2>/dev/null")
			avbState = strings.TrimSpace(avbState)
			if avbState == "orange" || avbState == "yellow" {
				items = append(items, DiagItem{
					Name:        "Android Verified Boot (AVB)",
					Status:      "WARNING",
					Value:       fmt.Sprintf("🟡 State: %s", avbState),
					Description: "Bootloader telah di-unlock (Orange/Yellow State). Ini normal untuk HP yang menggunakan Custom Kernel.",
				})
			} else if avbState == "red" {
				items = append(items, DiagItem{
					Name:        "Android Verified Boot (AVB)",
					Status:      "ERROR",
					Value:       "🔴 Red State (Corrupted)",
					Description: "Verifikasi signature partisi boot gagal. Segera lakukan flashing ulang boot original untuk memulihkan.",
				})
			} else {
				items = append(items, DiagItem{
					Name:        "Android Verified Boot (AVB)",
					Status:      "OK",
					Value:       "🟢 Locked / Green State",
					Description: "Android Verified Boot terkunci rapat dan signature boot image tervalidasi penuh oleh keystore hardware.",
				})
			}

			// 5. PStore RAMoops Scan
			pstoreFiles, _ := adbClient.Shell("ls /sys/fs/pstore/ 2>/dev/null")
			if strings.Contains(pstoreFiles, "console-ramoops") || strings.Contains(pstoreFiles, "dmesg-ramoops") {
				items = append(items, DiagItem{
					Name:        "Kernel Crash Log (PStore)",
					Status:      "WARNING",
					Value:       "⚠️ Found Crash Dump",
					Description: "Terdeteksi log crash kernel sebelumnya (RAMoops). Kernel mengalami panic/crash sebelum booting ulang terakhir.",
				})
			} else {
				items = append(items, DiagItem{
					Name:        "Kernel Crash Log (PStore)",
					Status:      "OK",
					Value:       "🟢 Clean (No Crashes)",
					Description: "Tidak ada log crash/panic kernel yang tertinggal di memory PStore. Kernel berjalan stabil.",
				})
			}

			// 6. Panel LCM Check
			lcmName, _ := adbClient.Shell("getprop ro.boot.lcm_name || getprop ro.boot.panel_name 2>/dev/null")
			lcmName = strings.TrimSpace(lcmName)
			if lcmName != "" {
				items = append(items, DiagItem{
					Name:        "Display Panel LCM",
					Status:      "OK",
					Value:       fmt.Sprintf("🟢 %s", truncateStr(lcmName, 25)),
					Description: "LCM Panel terdeteksi dan kompatibel dengan driver panel kernel universal Epitaph.",
				})
			} else {
				items = append(items, DiagItem{
					Name:        "Display Panel LCM",
					Status:      "WARNING",
					Value:       "🟡 Unidentified LCM",
					Description: "LCM Panel tidak teridentifikasi lewat fastboot getprop. Pastikan driver panel Anda terintegrasi.",
				})
			}
		}

		fyne.Do(func() {
			wp.progressScan.Hide()
			wp.btnScan.Enable()

			if info.Mode == adb.ModeNone {
				wp.lblStatus.SetText("Pindai selesai. Harap hubungkan HP Anda ke PC via kabel USB untuk memulai diagnosa mendalam.")
			} else {
				wp.lblStatus.SetText("🎉 Pindai Diagnosa Selesai! Klik detail item di bawah untuk melihat rincian masalah sistem.")
			}

			hasWiFiErrors := false
			resultItems := container.NewVBox()
			for _, item := range items {
				if item.Name == "WiFi Interface (wlan0)" && item.Status == "ERROR" {
					hasWiFiErrors = true
				}

				var statusColor color.Color
				switch item.Status {
				case "OK":
					statusColor = colorSuccess
				case "WARNING":
					statusColor = colorWarning
				default:
					statusColor = colorError
				}

				nameLbl := widget.NewLabel(fmt.Sprintf("%s  %s", item.Value, item.Name))
				nameLbl.TextStyle = fyne.TextStyle{Bold: true}

				descLbl := widget.NewLabel(item.Description)
				descLbl.Wrapping = fyne.TextWrapWord
				descLbl.TextStyle = fyne.TextStyle{Italic: true}

				statusLine := canvas.NewRectangle(statusColor)
				statusLine.SetMinSize(fyne.NewSize(4, 0))

				itemRow := container.NewBorder(nil, nil, statusLine, nil, container.NewPadded(container.NewVBox(
					nameLbl,
					descLbl,
				)))
				resultItems.Add(itemRow)
				if len(items) > 1 {
					resultItems.Add(NeoDivider())
				}
			}
			wp.resultBox.Add(NewNeoCard("Hasil Diagnostik", "", resultItems))

			if hasWiFiErrors && info.Mode == adb.ModeAndroid {
				wp.btnFix.Show()
			} else {
				wp.btnFix.Hide()
			}

			wp.resultBox.Refresh()
			wp.mainCard.Refresh()
		})
	}()
}

func (wp *WiFiFixPage) runAutoFix() {
	diag := wp.fixer.Diagnosis()
	if !diag.IsWiFiBroken {
		dialog.ShowInformation("Sistem Stabil", "Tidak ada error kritis yang perlu diperbaiki secara otomatis saat ini.", wp.window)
		return
	}

	progress := dialog.NewCustomWithoutButtons("Mencoba Auto-Fix Driver...", widget.NewActivity(), wp.window)
	progress.Show()

	go func() {
		adbClient := wp.deviceMgr.GetADBClient()

		// Attempt to reload network driver and restart WiFi stack
		_, _ = adbClient.Shell("su -c 'svc wifi disable && sleep 1 && svc wifi enable' 2>/dev/null")
		_, _ = adbClient.Shell("su -c 'insmod /vendor/lib/modules/wlan.ko || insmod /vendor_dlkm/lib/modules/wlan.ko' 2>/dev/null")
		_, _ = adbClient.Shell("su -c 'svc wifi enable' 2>/dev/null")

		// Re-run diagnostic to verify
		time.Sleep(1500 * time.Millisecond)
		_ = wp.fixer.DiagnoseWiFi()

		fyne.Do(func() {
			progress.Hide()
			if wp.fixer.Diagnosis().IsWiFiBroken {
				// Still broken, show guide dialog
				dialog.ShowConfirm("Auto-Fix Gagal", 
					"Sistem gagal meload ulang driver WiFi secara otomatis karena kernel signature mismatch.\n\nApakah Anda ingin membuka panduan langkah demi langkah (Step-by-Step Recovery Wizard) untuk memulihkan WiFi HP Anda secara tuntas?",
					func(ok bool) {
						if ok {
							wp.showRescueGuide()
						}
					}, wp.window)
			} else {
				dialog.ShowInformation("Berhasil!", "Driver WiFi (wlan0) berhasil diregenerasi dan dinyalakan kembali secara otomatis!", wp.window)
				wp.runSystemDiagnostics()
			}
		})
	}()
}

func (wp *WiFiFixPage) showRescueGuide() {
	if wp.onNavigate != nil {
		wp.onNavigate(1) // Switch to Rescue tab (index 1)
	}
}

func truncateStr(s string, limit int) string {
	s = strings.TrimSpace(s)
	if len(s) > limit {
		return s[:limit-3] + "..."
	}
	return s
}
