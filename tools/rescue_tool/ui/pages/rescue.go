package pages

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/naidrahiqa/epitaph_rescue/internal/adb"
	"github.com/naidrahiqa/epitaph_rescue/internal/rescue"
)



// RescuePage is the full rescue wizard UI
type RescuePage struct {
	wizard      *rescue.Wizard
	deviceMgr   *adb.DeviceManager
	window      fyne.Window
	onNavigate  func(index int)

	// Step indicators
	stepLabels  []*widget.Label
	stepIcons   []*canvas.Circle

	// Main content area
	statusLabel *widget.Label
	detailLabel *widget.Label
	actionBtn   *widget.Button
	progressBar *widget.ProgressBar
	resetBtn    *widget.Button

	mainBox     *fyne.Container
}

func (rp *RescuePage) SetOnNavigate(onNavigate func(index int)) {
	rp.onNavigate = onNavigate
}

// NewRescuePage creates the rescue wizard UI
func NewRescuePage(dm *adb.DeviceManager, w fyne.Window) *RescuePage {
	rp := &RescuePage{
		wizard:    rescue.NewWizard(dm),
		deviceMgr: dm,
		window:    w,
	}
	rp.wizard.SetOnUpdate(func() {
		rp.refreshUI()
	})
	rp.buildUI()
	return rp
}

func (rp *RescuePage) buildUI() {
	steps := []string{
		"1. Detect",
		"2. Flash",
		"3. Reboot",
		"4. Pull Log",
		"5. Analyze",
	}

	rp.stepLabels = make([]*widget.Label, len(steps))
	rp.stepIcons = make([]*canvas.Circle, len(steps))

	stepContainer := container.NewHBox(layout.NewSpacer())
	for i, name := range steps {
		circle := canvas.NewCircle(colorBorder)
		circle.StrokeColor = colorTextMuted
		circle.StrokeWidth = 2
		circle.Resize(fyne.NewSize(14, 14))
		rp.stepIcons[i] = circle

		lbl := widget.NewLabel(name)
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		rp.stepLabels[i] = lbl

		stepContainer.Add(container.NewHBox(circle, widget.NewLabel(" "), lbl))
		if i < len(steps)-1 {
			stepContainer.Add(widget.NewLabel("  →  "))
		}
	}
	stepContainer.Add(layout.NewSpacer())

	rp.statusLabel = NewNeoHeading("Rescue Wizard")
	rp.statusLabel.Wrapping = fyne.TextWrapWord

	initialDetail := "Klik 'Start Rescue' untuk memulai.\n\nAlur:\n1. Detect device status\n2. Flash stock boot.img (jika Fastboot)\n3. Wait reboot ke Android\n4. Pull crash log dari PStore\n5. Analyze log untuk root cause"
	rp.detailLabel = NewNeoLabel(initialDetail)

	rp.progressBar = widget.NewProgressBar()
	rp.progressBar.Hide()

	rp.actionBtn = widget.NewButtonWithIcon("Start Rescue", theme.MediaPlayIcon(), func() {
		rp.runNextStep()
	})
	rp.actionBtn.Importance = widget.HighImportance

	rp.resetBtn = widget.NewButtonWithIcon("Reset", theme.ViewRefreshIcon(), func() {
		rp.wizard.Reset()
		rp.detailLabel.SetText(initialDetail)
		rp.actionBtn.SetText("Start Rescue")
		rp.actionBtn.SetIcon(theme.MediaPlayIcon())
		rp.actionBtn.Enable()
		rp.progressBar.Hide()
		rp.refreshUI()
	})

	btnRow := container.NewHBox(
		layout.NewSpacer(),
		rp.resetBtn,
		rp.actionBtn,
		layout.NewSpacer(),
	)

	btnDirectFlash := NewNeoButton("Flash stock boot.img", theme.ConfirmIcon(), widget.LowImportance, func() {
		rp.directFlashBoot()
	})

	btnDirectPull := NewNeoButton("Tarik & Analisis Crash Log", theme.DownloadIcon(), widget.LowImportance, func() {
		rp.directPullLog()
	})

	btnRebootSysADB := NewNeoButton("Reboot System (ADB)", theme.ViewRefreshIcon(), widget.LowImportance, func() {
		rp.runADBCommand("reboot")
	})
	btnRebootBootloaderADB := NewNeoButton("Reboot Fastboot (ADB)", theme.SettingsIcon(), widget.LowImportance, func() {
		rp.runADBCommand("reboot bootloader")
	})
	btnRebootSysFB := NewNeoButton("Reboot System (Fastboot)", theme.ViewRefreshIcon(), widget.LowImportance, func() {
		rp.runFastbootCommand("reboot")
	})

	rebootRow := container.NewAdaptiveGrid(3,
		btnRebootSysADB,
		btnRebootBootloaderADB,
		btnRebootSysFB,
	)

	individualCard := NewNeoCard("Individual Actions",
		"Jalankan langkah rescue secara independen",
		container.NewVBox(
			container.NewAdaptiveGrid(2,
				btnDirectFlash,
				btnDirectPull,
			),
			NeoDivider(),
			NewNeoHeading("Kontrol Reboot Device:"),
			rebootRow,
		),
	)

	rp.mainBox = container.NewVBox(
		NewNeoCard("Rescue Wizard", "Step-by-step bootloop recovery", container.NewVBox(
			stepContainer,
			NeoDivider(),
			rp.statusLabel,
			rp.detailLabel,
			rp.progressBar,
			NeoDivider(),
			btnRow,
		)),
		individualCard,
	)
}

func (rp *RescuePage) Content() fyne.CanvasObject {
	return container.NewScroll(rp.mainBox)
}

func (rp *RescuePage) runNextStep() {
	state := rp.wizard.State()
	step, status, _ := state.GetStep()

	switch {
	case step == rescue.StepDetect && (status == rescue.StatusPending || status == rescue.StatusFailed):
		rp.actionBtn.Disable()
		go func() {
			rp.wizard.RunStep1Detect()
			fyne.Do(func() {
				rp.actionBtn.Enable()
				st := rp.wizard.State()
				s, ss, _ := st.GetStep()
				if ss == rescue.StatusSuccess {
					if s == rescue.StepDetect {
						info := rp.deviceMgr.Detect()
						if info.Mode == adb.ModeFastboot {
							rp.actionBtn.SetText("Flash Stock Boot")
							rp.actionBtn.SetIcon(theme.ConfirmIcon())
						} else {
							rp.actionBtn.SetText("Pull Crash Log")
							rp.actionBtn.SetIcon(theme.DownloadIcon())
						}
					}
				}
			})
		}()

	case step == rescue.StepDetect && status == rescue.StatusSuccess:
		info := rp.deviceMgr.Detect()
		if info.Mode == adb.ModeFastboot {
			// Need to select boot.img file
			rp.selectBootImage()
		} else {
			// Skip to pull log
			state.SetStep(rescue.StepPullLog, rescue.StatusPending, "")
			rp.actionBtn.SetText("Pull Crash Log")
			rp.actionBtn.SetIcon(theme.DownloadIcon())
			rp.actionBtn.Disable()
			go func() {
				rp.wizard.SetLogOutputDir(GetLogOutputDir())
				rp.wizard.RunStep4PullLog()
				fyne.Do(func() {
					rp.actionBtn.Enable()
					rp.actionBtn.SetText("Analyze Log")
					rp.actionBtn.SetIcon(theme.SearchIcon())
				})
			}()
		}

	case step == rescue.StepFlashBoot && status == rescue.StatusSuccess:
		rp.actionBtn.SetText("Waiting for Reboot...")
		rp.actionBtn.SetIcon(theme.SettingsIcon())
		rp.actionBtn.Disable()
		rp.progressBar.Show()
		rp.progressBar.SetValue(0)
		go func() {
			rp.wizard.RunStep3WaitReboot()
			fyne.Do(func() {
				rp.progressBar.Hide()
				rp.actionBtn.Enable()
				rp.actionBtn.SetText("Pull Crash Log")
				rp.actionBtn.SetIcon(theme.DownloadIcon())
			})
		}()

	case step == rescue.StepWaitReboot && status == rescue.StatusSuccess:
		rp.actionBtn.Disable()
		go func() {
			rp.wizard.SetLogOutputDir(GetLogOutputDir())
			rp.wizard.RunStep4PullLog()
			fyne.Do(func() {
				rp.actionBtn.Enable()
				rp.actionBtn.SetText("Analyze Log")
				rp.actionBtn.SetIcon(theme.SearchIcon())
			})
		}()

	case step == rescue.StepPullLog && status == rescue.StatusSuccess:
		rp.actionBtn.Disable()
		go func() {
			rp.wizard.RunStep5Analyze()
			fyne.Do(func() {
				rp.actionBtn.SetText("Buka Folder Log")
				rp.actionBtn.SetIcon(theme.FolderIcon())
				rp.actionBtn.Enable()
				rp.showAnalysisResults()
			})
		}()

	case step == rescue.StepComplete:
		OpenFolderInExplorer(GetLogOutputDir())

	default:
		// Retry current step
		rp.actionBtn.SetText("Retry")
		rp.actionBtn.SetIcon(theme.ViewRefreshIcon())
	}
}

func (rp *RescuePage) selectBootImage() {
	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		reader.Close()
		path := reader.URI().Path()

		rp.actionBtn.Disable()
		go func() {
			rp.wizard.RunStep2Flash(path)
			fyne.Do(func() {
				rp.actionBtn.Enable()

				st := rp.wizard.State()
				_, ss, _ := st.GetStep()
				if ss == rescue.StatusSuccess {
					rp.actionBtn.SetText("Wait Reboot")
					rp.actionBtn.SetIcon(theme.SettingsIcon())
				}
			})
		}()
	}, rp.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".img"}))
	fd.Show()
}

func (rp *RescuePage) showAnalysisResults() {
	state := rp.wizard.State()
	result := state.GetAnalysisResult()
	logPath := state.GetLogFilePath()

	if result == nil {
		return
	}

	detail := fmt.Sprintf("📊 Analisis Selesai!\n\n")
	detail += fmt.Sprintf("📄 Log disimpan di: %s\n", logPath)
	detail += fmt.Sprintf("📏 Total baris: %d\n", result.TotalLines)
	detail += fmt.Sprintf("🔍 Pattern match: %d\n\n", len(result.Matches))

	if len(result.TopIssues) > 0 {
		detail += "🔝 Top Issues:\n"
		detail += "━━━━━━━━━━━━━━━━━━━━━━━━\n"
		for i, issue := range result.TopIssues {
			severity := "ℹ️"
			switch issue.Severity {
			case 2: // CRITICAL
				severity = "🔴"
			case 1: // WARNING
				severity = "🟡"
			}
			detail += fmt.Sprintf("\n%s #%d [%s] %s\n", severity, i+1, issue.Category, issue.Diagnosis)
			detail += fmt.Sprintf("   💡 %s\n", issue.ActionHint)
			detail += fmt.Sprintf("   📍 Pertama muncul di baris %d (%d kali)\n", issue.FirstLine, issue.Count)
		}
	} else {
		detail += "✅ Tidak ada error pattern yang terdeteksi di log ini.\n"
	}

	rp.detailLabel.SetText(detail)
	rp.mainBox.Refresh()
}

func (rp *RescuePage) refreshUI() {
	state := rp.wizard.State()
	step, status, msg := state.GetStep()

	rp.statusLabel.SetText(fmt.Sprintf("Step: %s", step.String()))
	rp.detailLabel.SetText(msg)

	// Update step indicators
	for i := rescue.StepDetect; i <= rescue.StepAnalyze; i++ {
		idx := int(i)
		if idx >= len(rp.stepIcons) {
			break
		}

		st := state.StepStatuses[i]
		switch st {
		case rescue.StatusSuccess:
			rp.stepIcons[idx].FillColor = colorSuccess
		case rescue.StatusRunning:
			rp.stepIcons[idx].FillColor = colorSecondary
		case rescue.StatusFailed:
			rp.stepIcons[idx].FillColor = colorError
		case rescue.StatusSkipped:
			rp.stepIcons[idx].FillColor = colorTextMuted
		default:
			rp.stepIcons[idx].FillColor = colorBorder
		}
		rp.stepIcons[idx].Refresh()
	}

	_ = status
	rp.mainBox.Refresh()
}

func (rp *RescuePage) directFlashBoot() {
	info := rp.deviceMgr.Detect()
	if info.Mode != adb.ModeFastboot {
		dialog.ShowError(fmt.Errorf("device tidak terdeteksi di Fastboot Mode. Pastikan HP berada di Fastboot Mode sebelum flash."), rp.window)
		return
	}

	fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		reader.Close()
		path := reader.URI().Path()

		// Validate file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			dialog.ShowError(fmt.Errorf("file boot.img tidak ditemukan!"), rp.window)
			return
		}

		// Validate magic bytes (Android boot image starts with "ANDROID!")
		f, err := os.Open(path)
		if err != nil {
			dialog.ShowError(fmt.Errorf("gagal membuka file: %v", err), rp.window)
			return
		}
		magic := make([]byte, 8)
		n, _ := f.Read(magic)
		f.Close()

		if n < 8 || string(magic) != "ANDROID!" {
			dialog.ShowError(fmt.Errorf("file bukan boot.img yang valid! (Magic bytes tidak cocok — harus 'ANDROID!')"), rp.window)
			return
		}

		progress := dialog.NewCustomWithoutButtons("Flashing...", widget.NewActivity(), rp.window)
		progress.Show()

		go func() {
			fb := rp.deviceMgr.GetFastbootClient()
			output, err := fb.Flash("boot", path)
			progress.Hide()

			if err != nil {
				dialog.ShowError(fmt.Errorf("flash gagal: %v\nOutput: %s", err, output), rp.window)
			} else {
				dialog.ShowInformation("Berhasil!", fmt.Sprintf("Flash boot.img berhasil dilakukan!\n\nOutput:\n%s", output), rp.window)
			}
		}()
	}, rp.window)

	fd.SetFilter(storage.NewExtensionFileFilter([]string{".img"}))
	fd.Show()
}

func (rp *RescuePage) directPullLog() {
	info := rp.deviceMgr.Detect()
	if info.Mode != adb.ModeAndroid {
		dialog.ShowError(fmt.Errorf("device tidak terdeteksi di mode Android (ADB). Aktifkan USB Debugging."), rp.window)
		return
	}

	progress := dialog.NewCustomWithoutButtons("Menarik log...", widget.NewActivity(), rp.window)
	progress.Show()

	go func() {
		adbClient := rp.deviceMgr.GetADBClient()
		pstorePaths := []string{
			"/sys/fs/pstore/console-ramoops-0",
			"/sys/fs/pstore/dmesg-ramoops-0",
			"/sys/fs/pstore/console-ramoops",
			"/proc/last_kmsg",
		}

		var content string
		var matchedPath string

		for _, path := range pstorePaths {
			out, err := adbClient.Shell(fmt.Sprintf("cat %s 2>/dev/null || su -c 'cat %s' 2>/dev/null", path, path))
			if err == nil && strings.TrimSpace(out) != "" && len(out) > 100 {
				content = out
				matchedPath = path
				break
			}
		}

		fyne.Do(func() {
			progress.Hide()

			if content == "" {
				dialog.ShowError(fmt.Errorf("tidak ada crash log ditemukan di PStore. Pastikan HP mendukung PStore/RAMoops."), rp.window)
				return
			}

			outputDir := GetLogOutputDir()
			_ = os.MkdirAll(outputDir, 0755)
			timestamp := time.Now().Format("20060102_150405")
			tmpFile := filepath.Join(outputDir, fmt.Sprintf("pulled_ramoops_%s.txt", timestamp))
			_ = os.WriteFile(tmpFile, []byte(content), 0644)

			dialog.ShowConfirm("Berhasil!", fmt.Sprintf("Log berhasil ditarik dari %s\nDisimpan ke: %s\n\nIngin membuka dan menganalisis log ini di tab Log Analyzer sekarang?", matchedPath, tmpFile), func(ok bool) {
				if ok {
					if rp.onNavigate != nil {
						rp.onNavigate(2) // Switch to Log tab (index 2)
					}
				}
			}, rp.window)
		})
	}()
}

func (rp *RescuePage) runADBCommand(args string) {
	info := rp.deviceMgr.Detect()
	if info.Mode != adb.ModeAndroid {
		dialog.ShowError(fmt.Errorf("device tidak terdeteksi di mode Android (ADB)."), rp.window)
		return
	}
	
	progress := dialog.NewCustomWithoutButtons("Menjalankan perintah ADB...", widget.NewActivity(), rp.window)
	progress.Show()
	
	go func() {
		adbClient := rp.deviceMgr.GetADBClient()
		_, err := adbClient.Run(strings.Fields(args)...)
		fyne.Do(func() {
			progress.Hide()
			if err != nil {
				dialog.ShowError(fmt.Errorf("gagal menjalankan adb %s: %v", args, err), rp.window)
			} else {
				dialog.ShowInformation("Sukses", fmt.Sprintf("Berhasil mengirim perintah: adb %s", args), rp.window)
			}
		})
	}()
}

func (rp *RescuePage) runFastbootCommand(args string) {
	info := rp.deviceMgr.Detect()
	if info.Mode != adb.ModeFastboot {
		dialog.ShowError(fmt.Errorf("device tidak terdeteksi di mode Fastboot."), rp.window)
		return
	}
	
	progress := dialog.NewCustomWithoutButtons("Menjalankan perintah Fastboot...", widget.NewActivity(), rp.window)
	progress.Show()
	
	go func() {
		fbClient := rp.deviceMgr.GetFastbootClient()
		_, err := fbClient.Run(strings.Fields(args)...)
		fyne.Do(func() {
			progress.Hide()
			if err != nil {
				dialog.ShowError(fmt.Errorf("gagal menjalankan fastboot %s: %v", args, err), rp.window)
			} else {
				dialog.ShowInformation("Sukses", fmt.Sprintf("Berhasil mengirim perintah: fastboot %s", args), rp.window)
			}
		})
	}()
}
